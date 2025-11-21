package handlers

import (
	"videoapp/internal/generated/proto"

	"github.com/gofiber/fiber/v2"
)

func react(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(400)
	}

	_type, err := c.ParamsInt("type")
	if err != nil {
		return c.SendStatus(400)
	}

	targetType := c.QueryInt("target_type")

	if _, err := deps.Clients.Reactions.React(c.Context(), &proto.ReactRequest{Session: c.Cookies("session"), VideoId: int64(id), Type: int32(_type), TargetType: proto.TargetType(targetType)}); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}
func deleteReaction(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.SendStatus(400)
	}

	targetType := c.QueryInt("target_type")

	if _, err := deps.Clients.Reactions.Remove(c.Context(), &proto.RemoveRequest{Session: c.Cookies("session"), VideoId: int64(id), TargetType: proto.TargetType(targetType)}); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}
