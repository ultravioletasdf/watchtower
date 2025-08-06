package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"videoapp/proto"
	sqlc "videoapp/sql"
	"videoapp/utils"
	"videoapp/vips"

	"github.com/bwmarrin/snowflake"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var cfg utils.Config
var snowflakeNode *snowflake.Node
var executor *sqlc.Queries
var s3 *minio.Client
var rabbit Rabbit

type Rabbit struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queues     struct {
		AnalyseVideos      *amqp091.Queue
		TranscodeVideos    *amqp091.Queue
		GenerateThumbnails *amqp091.Queue
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
	cfg = utils.ParseConfig()

	var err error
	snowflakeNode, err = snowflake.NewNode(cfg.SnowflakeNode)
	if err != nil {
		log.Fatalf("Failed to create snowflake node: %v", err)
	}

	s3 = utils.ConnectS3(cfg)
	db := utils.ConnectDatabase(cfg)
	defer db.Close(context.Background())
	executor = db.Queries

	connectRabbit()
	defer rabbit.Close()

	vips.Startup(nil)
	defer vips.Shutdown()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
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

func connectRabbit() {
	conn, err := amqp091.Dial(cfg.AmqpUrl)
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
	transcodeQueue, err := rabbit.channel.QueueDeclare("transcode_video", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}
	thumbnailQueue, err := rabbit.channel.QueueDeclare("generate_thumbnails", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	rabbit.queues.AnalyseVideos = &analyseVideoQueue
	rabbit.queues.TranscodeVideos = &transcodeQueue
	rabbit.queues.GenerateThumbnails = &thumbnailQueue
}
