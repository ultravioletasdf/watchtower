package main

import (
	"context"
	"flag"
	"log"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	amqp "github.com/rabbitmq/amqp091-go"
)

var minioClient *minio.Client

func main() {
	envFile := flag.String("env", "", "File path to env file")
	flag.Parse()

	if envFile != nil && *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			panic(err)
		}
	}
	parseConfig()

	connectMinio()

	conn, err := amqp.Dial(config.AmqpUrl)
	if err != nil {
		log.Fatalf("Failed to dial rabbit: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare("analyse_video", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}
	var forever chan struct{}

	go func() {
		for d := range msgs {
			handleMessage(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func connectMinio() {
	client, err := minio.New(config.Minio.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(config.Minio.AccessKey, config.Minio.SecretKey, ""), Secure: false})
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}
	minioClient = client
	exists, err := minioClient.BucketExists(context.Background(), "staging")
	if err != nil {
		log.Fatalf("failed to verify bucket exists: %v", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(context.Background(), "staging", minio.MakeBucketOptions{}); err != nil {
			log.Fatalf("failed to create bucket: %v", err)
		}
		log.Println("Created bucket staging")
	}
}
