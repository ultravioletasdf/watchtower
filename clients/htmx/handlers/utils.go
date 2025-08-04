package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"
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
