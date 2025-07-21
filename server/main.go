package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"videoapp/proto"
	sqlc "videoapp/sql"

	"github.com/bwmarrin/snowflake"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var snowflakeNode *snowflake.Node
var executor *sqlc.Queries
var minioClient *minio.Client

var rabbit Rabbit

type Rabbit struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queues     struct {
		AnalyseVideos *amqp091.Queue
	}
}

func (r *Rabbit) Close() {
	defer r.connection.Close()
	defer r.channel.Close()
}

func main() {
	loadEnv := flag.String("env", "", "Specify whether to read .env")
	flag.Parse()
	if *loadEnv != "" {
		if err := godotenv.Load(*loadEnv); err != nil {
			log.Printf("Failed to load env: %v\n(continuing)\n", err)
		}
	}
	parseConfig()

	var err error
	snowflakeNode, err = snowflake.NewNode(config.SnowflakeNode)
	if err != nil {
		log.Fatalf("Failed to create snowflake node: %v", err)
	}

	connectDb()
	connectMinio()
	connectRabbit()
	defer rabbit.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterUsersServer(s, &userServer{})
	proto.RegisterSessionsServer(s, &sessionServer{})
	proto.RegisterVideosServer(s, &videoService{})
	proto.RegisterThumbnailsServer(s, &thumbnailsServer{})
	reflection.Register(s)

	log.Printf("server listening at %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func connectDb() {
	db, err := sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatalf("failed to connect to sqlite: %v", err)
	}
	executor = sqlc.New(db)
}
func connectMinio() {
	client, err := minio.New(config.Minio.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(config.Minio.AccessKey, config.Minio.SecretKey, ""), Secure: false})
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}
	minioClient = client

	buckets := []string{"staging", "thumbnails", "staging-thumbnails"}

	for _, bucket := range buckets {
		exists, err := minioClient.BucketExists(context.Background(), bucket)
		if err != nil {
			log.Fatalf("failed to verify bucket exists: %v", err)
		}
		if !exists {
			if err := minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
				log.Fatalf("failed to create bucket: %v", err)
			}
			log.Printf("Created bucket %s\n", bucket)
		}
		policyFile := fmt.Sprintf("policies/%s.json", bucket)
		policy, err := os.ReadFile(policyFile)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			log.Fatalf("failed to get bucket policy %s: %v", policyFile, err)
		}
		err = minioClient.SetBucketPolicy(context.Background(), bucket, string(policy))
		if err != nil {
			log.Fatalf("failed to apply bucket policy %s: %v", policyFile, err)
		}
	}
}
func connectRabbit() {
	conn, err := amqp091.Dial(config.AmqpUrl)
	if err != nil {
		log.Fatalf("Failed to connect to rabbitmq: %v", err)
	}
	rabbit.connection = conn

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open rabbitmq channel: %v", err)
	}
	rabbit.channel = channel

	analyseVideoQueue, err := rabbit.channel.QueueDeclare("analyse_video", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
	rabbit.queues.AnalyseVideos = &analyseVideoQueue
}
