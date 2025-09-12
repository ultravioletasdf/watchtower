package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"

	"videoapp/cmd/web/frontend"
	"videoapp/internal/proto"
)

func follow(c *fiber.Ctx) error {
	idString := c.Params("id")
	id, err := strconv.ParseInt(idString, 10, 0)
	if err != nil {
		return c.Status(400).SendString("Invalid id")
	}
	switch c.Method() {
	case "POST":
		_, err = deps.Clients.Users.Follow(ctx, &proto.FollowRequest{Session: c.Cookies("session"), IdToFollow: id})
		if err != nil {
			err := status.Convert(err)
			return c.Status(500).SendString(err.Message())
		}
		return Render(c, frontend.ButtonUnfollow(id))
	case "DELETE":
		_, err = deps.Clients.Users.Unfollow(ctx, &proto.FollowRequest{Session: c.Cookies("session"), IdToFollow: id})
		if err != nil {
			err := status.Convert(err)
			return c.Status(500).SendString(err.Message())
		}
		return Render(c, frontend.ButtonFollow(id))
	default:
		return c.SendStatus(fiber.StatusMethodNotAllowed)
	}
}
