package queues

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	"videoapp/internal/generated/proto"
)

var ctx = context.Background()

func DownloadVideo(s3 *minio.Client, d amqp091.Delivery, message *proto.AnalyseVideoMessage) error {
	uploadId := strconv.FormatInt(message.UploadId, 10)

	log.Printf("Starting download for upload %s\n", uploadId)
	if err := s3.FGetObject(ctx, "staging", uploadId, "videos/"+uploadId, minio.GetObjectOptions{}); err != nil {
		log.Printf("Failed to get object: %v", err)
		if err.Error() == "The specified key does not exist." {
			d.Reject(false)
			return err
		}
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return err
	}
	log.Printf("Download completed for %s\n", uploadId)
	return nil
}

// Deletes all directors videos/{uploadId} and results/{uploadId}
func Cleanup(uploadId int64) {
	paths := []string{fmt.Sprintf("results/%d", uploadId), fmt.Sprintf("videos/%d", uploadId)}
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			log.Printf("failed to clean up %s: %v\n", path, err)
		}
	}
}

// Deletes videos/ and results/* recursively
func CleanupAll() {
	paths := []string{"results", "videos"}
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			log.Printf("failed to clean up %s: %v\n", path, err)
		}
	}
}

// Splits a video into frames
func Split(d amqp091.Delivery, uploadId string, vf string) error {
	err := os.MkdirAll("results/"+uploadId, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create directories: %v\n", err)
		return err
	}
	output := "results/" + uploadId + "/frame_%03d.jpeg"
	if err := ffmpeg.Input("videos/"+uploadId, ffmpeg.KwArgs{}).Output(output, ffmpeg.KwArgs{"vf": vf, "vsync": "vfr"}).OverWriteOutput().Run(); err != nil {
		log.Printf("Failed to split into frames: %v\n", err)
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return err
	}
	log.Printf("Handled upload %s", uploadId)
	return nil
}
