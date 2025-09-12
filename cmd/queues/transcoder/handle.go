package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/grafov/m3u8"
	"github.com/rabbitmq/amqp091-go"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	protobuf "google.golang.org/protobuf/proto"
	"gopkg.in/vansante/go-ffprobe.v2"

	"videoapp/internal/proto"
	"videoapp/internal/queues"
	sqlc "videoapp/sql"
)

var STANDARD_HEIGHTS = []int{360, 480, 720, 1080, 1440, 2160}
var MIN_ASPECT_RATIO = 0.5
var MAX_ASPECT_RATIO = 3.0

type Resolution struct {
	Height int
	Width  int
}

func handleMessage(d amqp091.Delivery) {
	var message proto.AnalyseVideoMessage
	if err := protobuf.Unmarshal(d.Body, &message); err != nil {
		log.Printf("Failed to parse body: %v\n", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return
	}

	log.Printf("Received a message: %s\n", &message)
	if err := queues.DownloadVideo(s3, d, &message); err != nil {
		return
	}

	if err := queries.UpdateVideoStage(ctx, sqlc.UpdateVideoStageParams{ID: message.VideoId, Stage: int32(proto.Stage_Processing)}); err != nil {
		log.Printf("failed to update stage: %v\n", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return
	}

	if err := os.MkdirAll(fmt.Sprintf("results/%d", message.UploadId), os.ModePerm); err != nil {
		log.Printf("Failed to create directories, attempting to continue: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fname := fmt.Sprintf("videos/%d", message.UploadId)
	data, err := ffprobe.ProbeURL(ctx, fname)
	if err != nil {
		log.Printf("failed to probe %d: %v\n", message.UploadId, err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return
	}
	videoStream := data.FirstVideoStream()
	if videoStream == nil {
		log.Printf("rejecting video %d, had no video stream\n", message.UploadId)
		d.Reject(false)
		return
	}
	ratio := float64(videoStream.Width) / float64(videoStream.Height)
	if MIN_ASPECT_RATIO > ratio {
		log.Printf("rejecting video %d, aspect ratio %f is less than minimum %f\n", message.UploadId, ratio, MIN_ASPECT_RATIO)
		d.Reject(false)
		return
	}
	audio := data.FirstAudioStream()
	fmt.Println(audio)
	resolutions := []Resolution{}
	for _, STD_HEIGHT := range STANDARD_HEIGHTS {
		if videoStream.Height >= STD_HEIGHT {
			width := int(math.Round(float64(STD_HEIGHT) * (ratio)))
			// Make even to allow for chroma subsampling
			if width%2 != 0 {
				width++
			}
			resolutions = append(resolutions, Resolution{Height: STD_HEIGHT, Width: (width)})
		}
	}

	fmt.Println(resolutions)

	basePath := "results/" + strconv.FormatInt(message.UploadId, 10)
	for _, res := range resolutions {
		outputPath := fmt.Sprintf("%s/%dx%d.m3u8", basePath, res.Width, res.Height)
		segmentFilename := fmt.Sprintf("%s/%dx%d_%%d.ts", basePath, res.Width, res.Height)
		inputArgs := ffmpeg_go.KwArgs{}
		outputArgs := ffmpeg_go.KwArgs{
			"vf":                   fmt.Sprintf("scale=w=%d:h=%d", res.Width, res.Height),
			"c:v":                  "libx265",
			"tag:v":                "hvc1",
			"c:a":                  "aac",
			"ar":                   "48000",
			"b:a":                  "192k",
			"channel_layout":       "stereo",
			"ac":                   "2",
			"start_number":         0,
			"hls_time":             10,
			"hls_list_size":        0,
			"hls_segment_filename": segmentFilename,
			"f":                    "hls",
		}

		if cfg.TranscodeNvidia {
			inputArgs["hwaccel"] = "cuda"
			inputArgs["hwaccel_output_format"] = "cuda"

			outputArgs["c:v"] = "hevc_nvenc"
			outputArgs["vf"] = fmt.Sprintf("scale_cuda=w=%d:h=%d", res.Width, res.Height)
		}
		if err := ffmpeg_go.Input(fname, inputArgs).Output(outputPath, outputArgs).Run(); err != nil {
			log.Printf("failed to convert %d to hls: %v\n", message.UploadId, err)
			if err := d.Reject(true); err != nil {
				log.Printf("Failed to requeue: %v\n", err)
			}
			return
		}
	}
	log.Printf("finished processing %d\n", message.UploadId)

	master := m3u8.NewMasterPlaylist()
	for _, res := range resolutions {
		playlistFile := fmt.Sprintf("%dx%d.m3u8", res.Width, res.Height)
		max, avg, err := GetBitrate(basePath, res)
		if err != nil {
			log.Printf("failed to get bitrate: %v\n", err)
			if err := d.Reject(true); err != nil {
				log.Printf("Failed to requeue: %v\n", err)
			}
			return
		}
		resString := fmt.Sprintf("%dx%d", res.Width, res.Height)
		master.Append(playlistFile, nil, m3u8.VariantParams{AverageBandwidth: uint32(avg), Bandwidth: uint32(max), Resolution: resString})
	}

	f, err := os.Create(filepath.Join(basePath, "master.m3u8"))
	if err != nil {
		log.Printf("failed to create master: %v\n", err)
		return
	}
	defer f.Close()
	if _, err = master.Encode().WriteTo(f); err != nil {
		log.Printf("failed to write to master: %v\n", err)
		return
	}

	if err := UploadDir(basePath); err != nil {
		log.Printf("failed to upload files: %v\n", err)
		return
	}
	log.Printf("finished uploading %d\n", message.UploadId)

	queues.Cleanup(message.UploadId)
	if err := queries.UpdateVideoStage(context.Background(), sqlc.UpdateVideoStageParams{ID: message.VideoId, Stage: int32(proto.Stage_Processed)}); err != nil {
		log.Printf("failed to update stage: %v\n", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return
	}
	if err := d.Ack(false); err != nil {
		log.Printf("failed to ack: %v\n", err)
	}
}
