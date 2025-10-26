package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/bwmarrin/snowflake"
	gorseClient "github.com/gorse-io/gorse-go"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"videoapp/internal/generated/proto"
	"videoapp/internal/generated/sqlc"
	"videoapp/internal/generated/vips"
	"videoapp/internal/utils"

	_ "github.com/mattn/go-sqlite3"
)

var cfg utils.Config
var snowflakeNode *snowflake.Node
var executor *sqlc.Queries
var s3 *minio.Client
var privateKey *ecdsa.PrivateKey
var rabbit Rabbit
var gorse *gorseClient.GorseClient

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

	generateOrReadKeys()
	s3 = utils.ConnectS3(cfg)
	db := utils.ConnectDatabase(cfg)
	defer db.Close(context.Background())
	executor = db.Queries
	gorse = gorseClient.NewGorseClient(cfg.Gorse.Address, cfg.Gorse.Key)
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
	proto.RegisterVideosServer(s, &videoServer{})
	proto.RegisterThumbnailsServer(s, &thumbnailsServer{})
	proto.RegisterReactionsServer(s, &reactionsServer{})
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

func generateOrReadKeys() {
	bytes, err := os.ReadFile("./private_key.pem")
	if errors.Is(err, os.ErrNotExist) {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(err)
		}

		privBytes, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			panic(err)
		}
		privPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
		if err := os.WriteFile("./private_key.pem", privPem, 0600); err != nil {
			panic(err)
		}
		privateKey = priv
		fmt.Println("Created key")
	} else if err == nil {
		block, _ := pem.Decode(bytes)
		if block == nil || block.Type != "EC PRIVATE KEY" {
			panic("Failed to decode private key")
		}
		priv, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			panic(err)
		}
		privateKey = priv
	} else {
		panic("There was an error reading the private key: " + err.Error())
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if err := os.WriteFile("./public_key.pem", pubPem, 0644); err != nil {
		panic(err)
	}
}
