package handlers

import (
	"github.com/gofiber/fiber/v2"

	"videoapp/cmd/web/frontend"
)

func settings(c *fiber.Ctx) error {
	u := getUser(c)
	if u == nil {
		return c.Redirect("/sign/in")
	}

	return Render(c, frontend.Settings(u))
}
