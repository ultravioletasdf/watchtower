package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"videoapp/clients/htmx/frontend"
	"videoapp/proto"
	"videoapp/server/common"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"
)

func profile(c *fiber.Ctx) error {
	res, err := deps.Clients.Videos.GetUserVideos(ctx, &proto.GetUserVideosRequest{Session: c.Cookies("session")})
	if err == nil {
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
func getStages(c *fiber.Ctx) error {
	var ids []string
	if err := c.BodyParser(&ids); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	idInts := make([]int64, len(ids))
	for i, id := range ids {
		var err error
		idInts[i], err = strconv.ParseInt(id, 10, 0)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
	}
	res, err := deps.Clients.Videos.GetStages(ctx, &proto.VideosGetStagesRequest{Ids: idInts})
	if shouldReturn := unwrapGrpcError(c, err, 400); shouldReturn {
		return nil
	}
	return c.SendString(res.Result)
}
