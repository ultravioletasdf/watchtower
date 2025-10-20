package handlers

import (
	"strconv"
	"videoapp/internal/generated/proto"

	"github.com/gofiber/fiber/v2"
)

func react(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	_type := c.Params("type")
	typeInt, err := strconv.ParseInt(_type, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	if _, err := deps.Clients.Reactions.React(c.Context(), &proto.ReactRequest{Session: c.Cookies("session"), VideoId: idInt, Type: int32(typeInt)}); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}
func deleteReaction(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}

	if _, err := deps.Clients.Reactions.Remove(c.Context(), &proto.RemoveRequest{Session: c.Cookies("session"), VideoId: idInt}); err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendStatus(200)
}
