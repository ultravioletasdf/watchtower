package main

import (
	"context"
	"flag"
	"log"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	amqp "github.com/rabbitmq/amqp091-go"

	"videoapp/internal/queues"
	"videoapp/internal/utils"
	sqlc "videoapp/sql"
)

var cfg utils.Config
var s3 *minio.Client
var queries *sqlc.Queries

func main() {
	queues.CleanupAll()

	envFile := flag.String("env", "", "File path to env file")
	flag.Parse()

	if envFile != nil && *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			panic(err)
		}
	}
	cfg = utils.ParseConfig()

	s3 = utils.ConnectS3(cfg)
	db := utils.ConnectDatabase(cfg)
	defer db.Close(context.Background())
	queries = db.Queries

	conn, err := amqp.Dial(cfg.AmqpUrl)
	if err != nil {
		log.Fatalf("Failed to dial rabbit: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare("transcode_video", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
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
