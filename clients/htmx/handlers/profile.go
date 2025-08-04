package handlers

import (
	"errors"
	"fmt"
	"videoapp/clients/htmx/frontend"
	"videoapp/proto"
	"videoapp/server/common"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"
)

func profile(c *fiber.Ctx) error {
	res, err := deps.Clients.Videos.GetUserVideos(ctx, &proto.GetUserVideosRequest{Session: c.Cookies("session")})
	if err == nil {
		fmt.Println(res)
		return Render(c, frontend.Profile(res.Videos))
	}
	status, ok := status.FromError(err)
	if ok && errors.Is(status.Err(), common.ErrSessionNotFound) || errors.Is(status.Err(), common.ErrSessionWrongSize) {
		c.ClearCookie("session")
	} else {
		fmt.Printf("failed to get user: %v\n", err)
	}
	return c.Redirect("/sign/in")
}
