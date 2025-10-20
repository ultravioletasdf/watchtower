package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"videoapp/cmd/web/handlers"
	"videoapp/internal/generated/proto"
	"videoapp/internal/utils"
)

func main() {
	var deps handlers.Dependencies
	deps.Config = utils.ParseConfig()

	serverAddress := fmt.Sprintf("%s:%d", deps.Config.Server.Ip, deps.Config.Server.Port)
	client, err := grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to server %s: %v\n", serverAddress, err)
	}
	defer client.Close()
	deps.Clients.Users = proto.NewUsersClient(client)
	deps.Clients.Sessions = proto.NewSessionsClient(client)
	deps.Clients.Videos = proto.NewVideosClient(client)
	deps.Clients.Thumbnails = proto.NewThumbnailsClient(client)
	deps.Clients.Reactions = proto.NewReactionsClient(client)
	app := fiber.New(fiber.Config{Prefork: deps.Config.Web.Prefork})

	handlers.Add(app, deps)

	panic(app.Listen(deps.Config.Web.ListenAddress))
}
