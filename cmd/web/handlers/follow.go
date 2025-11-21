package handlers

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"

	"videoapp/cmd/web/frontend"
	"videoapp/internal/generated/proto"
)

func follow(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).SendString("Invalid id")
	}
	switch c.Method() {
	case "POST":
		_, err = deps.Clients.Users.Follow(ctx, &proto.FollowRequest{Session: c.Cookies("session"), IdToFollow: int64(id)})
		if err != nil {
			err := status.Convert(err)
			return c.Status(500).SendString(err.Message())
		}
		return Render(c, frontend.ButtonUnfollow(int64(id)))
	case "DELETE":
		_, err = deps.Clients.Users.Unfollow(ctx, &proto.FollowRequest{Session: c.Cookies("session"), IdToFollow: int64(id)})
		if err != nil {
			err := status.Convert(err)
			return c.Status(500).SendString(err.Message())
		}
		return Render(c, frontend.ButtonFollow(int64(id)))
	default:
		return c.SendStatus(fiber.StatusMethodNotAllowed)
	}
}
