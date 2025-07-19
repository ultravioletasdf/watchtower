package main

import (
	"context"
	"log"
	"strconv"
	"videoapp/proto"

	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
	protobuf "google.golang.org/protobuf/proto"
)

var ctx = context.Background()

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
	log.Printf("Starting download for upload %s\n", uploadId)
	if err := minioClient.FGetObject(ctx, "staging", uploadId, "videos/"+uploadId, minio.GetObjectOptions{}); err != nil {
		if err := d.Reject(true); err != nil {
			log.Printf("Failed to requeue: %v\n", err)
		}
		return
	}
	log.Printf("Download completed for %s\n", uploadId)

}
