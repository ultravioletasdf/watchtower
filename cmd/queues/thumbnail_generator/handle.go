package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
	protobuf "google.golang.org/protobuf/proto"
	"gopkg.in/gographics/imagick.v3/imagick"

	"videoapp/internal/generated/proto"
	"videoapp/internal/queues"
)

const (
	columns = 10
	height  = 90
)

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
	uploadId := strconv.FormatInt(message.UploadId, 10)
	if err := queues.Split(d, uploadId, "fps=1/3,scale=-1:90"); err != nil {
		return
	}
	numberOfFiles, err := genStoryBoard(d, uploadId)
	if err != nil {
		return
	}
	if err := genVtt(d, uploadId, numberOfFiles); err != nil {
		return
	}
	if err := upload(uploadId); err != nil {
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
	}
	if err := d.Ack(false); err != nil {
		log.Printf("failed to ack: %v\n", err)
	}
	log.Printf("Completed thumbnail generation for video %d, upload %d\n", message.VideoId, message.UploadId)
	queues.Cleanup(message.UploadId)
}
func genStoryBoard(d amqp091.Delivery, uploadId string) (int, error) {
	files, err := os.ReadDir("results/" + uploadId)
	if err != nil {
		log.Printf("failed to read results directory: %v\n", err.Error())
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return 0, err
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		err := mw.ReadImage(path.Join("results", uploadId, file.Name()))
		if err != nil {
			log.Printf("failed to read image: %v\n", err)
			if err := d.Reject(true); err != nil {
				log.Printf("Failed to requeue: %v\n", err)
			}
			return 0, err
		}
	}
	dw := imagick.NewDrawingWand()
	defer dw.Destroy()

	sprite := mw.MontageImage(dw, "10x", "x90+0+0>", imagick.MONTAGE_MODE_CONCATENATE, "")
	if sprite == nil {
		log.Println("failed to create montage")
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return 0, err
	}
	defer sprite.Destroy()

	sprite.SetImageFormat("webp")
	err = sprite.WriteImage(fmt.Sprintf("results/%s/storyboard.webp", uploadId))
	if err != nil {
		log.Printf("failed to write storyboard: %v\n", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return 0, err
	}
	log.Println("created storyboard")
	return len(files), nil
}
func genVtt(d amqp091.Delivery, uploadId string, numberOfFiles int) error {
	vtt, err := os.Create("results/" + uploadId + "/thumbnails.vtt")
	if err != nil {
		log.Printf("failed to open thumbnail file: %v\n", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return err
	}

	width, err := findWidth(uploadId)
	if err != nil {
		log.Printf("failed to find image width: %v\n", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return err
	}
	vtt.WriteString("WEBVTT\n")

	var t time.Time
	var row, col int
	for range numberOfFiles {
		if col >= 10 {
			col = 0
			row++
		}
		formattedTime := t.Format("15:04:05.000")
		t = t.Add(3 * time.Second)
		secondTime := t.Format("15:04:05.000")
		fmt.Fprintf(vtt, "\n%s --> %s\nstoryboard.webp#xywh=%d,%d,%d,%d\n", formattedTime, secondTime, col*width, row*90, width, 90)
		col++
	}
	return nil
}
func findWidth(uploadId string) (int, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ReadImage(path.Join("results", uploadId, "frame_001.jpeg"))
	if err != nil {
		return 0, err
	}
	return int(mw.GetImageWidth()), nil
}

type file struct {
	Name        string
	ContentType string
}

func upload(uploadId string) error {
	files := []file{file{Name: "storyboard.webp", ContentType: "image/webp"}, file{Name: "thumbnails.vtt", ContentType: "text/vtt"}}
	for _, file := range files {
		_, err := s3.FPutObject(context.Background(), "videos", fmt.Sprintf("%s/%s", uploadId, file.Name), path.Join("results", uploadId, file.Name), minio.PutObjectOptions{ContentType: file.ContentType})
		if err != nil {
			log.Printf("failed to upload %s: %v\n", file.Name, err)
			return err
		}
	}
	return nil
}
