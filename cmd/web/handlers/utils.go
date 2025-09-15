package handlers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"

	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
)

// Responds to the client with the original message from the grpc server, or an unknown error message if isn't from the grpc server
func unwrapGrpcError(c *fiber.Ctx, err error, statusCode int) bool {
	if err != nil {
		status, grpcError := status.FromError(err)
		if grpcError {
			c.Status(statusCode).SendString(status.Message())
			return true
		}
		c.Status(statusCode).SendString(fmt.Sprintf("Unknown Error: %s", err.Error()))
		return true
	}
	return false
}
func getUser(c *fiber.Ctx) *proto.User {
	session := c.Cookies("session")
	if session == "" {
		return nil
	}
	res, err := deps.Clients.Sessions.GetUser(ctx, &proto.Session{Token: session})
	if err != nil {
		if errors.Is(err, common.ErrSessionNotFound) {
			c.ClearCookie("session")
		} else {
			fmt.Printf("Failed to get user: %v\n", err)
		}
	}
	return res
}
