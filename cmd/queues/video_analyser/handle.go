package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/corona10/goimagehash"
	"github.com/rabbitmq/amqp091-go"
	protobuf "google.golang.org/protobuf/proto"

	"videoapp/internal/generated/proto"
	"videoapp/internal/generated/sqlc"
	"videoapp/internal/queues"
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
	uploadId := strconv.FormatInt(message.UploadId, 10)

	if err := queues.DownloadVideo(s3, d, &message); err != nil {
		return
	}

	if err := queues.Split(d, uploadId, "fps=1,scale=-1:300"); err != nil {
		return
	}
	if err := dedupe(uploadId); err != nil {
		log.Printf("Failed to dedupe: %v", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return
	}
	nsfw := findNsfw(uploadId)
	if nsfw {
		if err := queries.UpdateVideoStage(context.Background(), sqlc.UpdateVideoStageParams{ID: message.VideoId, Stage: int32(proto.Stage_FlaggedForNudity)}); err != nil {
			log.Printf("Failed to flag %d: %v\n", message.VideoId, err)
		}
	}

	if err := d.Ack(false); err != nil {
		log.Printf("failed to ack: %v\n", err)
	}
	queues.Cleanup(message.UploadId)
}

func dedupe(uploadId string) error {
	hashes := make(map[uint64]string)
	thresh := 5
	count := 0
	err := filepath.WalkDir("results/"+uploadId, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()
		img, err := jpeg.Decode(file)
		if err != nil {
			return err
		}
		hash, err := goimagehash.PerceptionHash(img)
		if err != nil {
			return err
		}

		for h, existing := range hashes {
			dist, err := goimagehash.NewImageHash(h, goimagehash.PHash).Distance(hash)
			if err != nil {
				log.Printf("Failed to get distance: %v\n", err)
				continue
			}
			if dist <= thresh {
				if err := os.Remove(existing); err != nil {
					log.Printf("Failed to remove similar files, %v\n", err)
				}
				count++
				delete(hashes, h)
			}
		}
		hashes[hash.GetHash()] = path
		return nil
	})
	log.Printf("Cleaned up %d similar images\n", count)
	return err
}
func findNsfw(uploadId string) bool {
	files, err := os.ReadDir("results/" + uploadId)
	if err != nil {
		panic(err)
	}
	var suspicousCount int

	for _, file := range files {
		if suspicousCount >= 3 {
			log.Println("Reached suspicous image threshold")
			return true
		}
		if file.IsDir() {
			continue
		}
		img := predictor.NewImage(fmt.Sprintf("results/%s/%s", uploadId, file.Name()), 3)
		res := predictor.Predict(img)

		if res.Hentai > 0.99 || res.Porn > 0.95 || res.Sexy > 0.95 {
			log.Println("Super suspicous image")
			return true
		}
		if res.Hentai > 0.95 || res.Porn > 0.9 || res.Sexy > 0.9 {
			log.Printf("Suspicious image %s, H: %.2f%%; P: %.2f%%; S %.2f%%;", file.Name(), res.Hentai*100, res.Porn*100, res.Sexy*100)
			suspicousCount++
		}
	}
	fmt.Printf("%d suspicous images for %s\n", suspicousCount, uploadId)
	return false
}
