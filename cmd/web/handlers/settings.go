package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"

	"videoapp/cmd/web/frontend"
	common "videoapp/internal/errors"
	"videoapp/internal/generated/proto"
)

func settings(c *fiber.Ctx) error {
	u := getUser(c)
	if u == nil {
		return c.Redirect("/sign/in")
	}

	return Render(c, frontend.Settings(u))
}
func putProfile(c *fiber.Ctx) error {
	session := c.Cookies("session")
	if session == "" {
		return c.Status(401).SendString("Please sign in to continue")
	}
	displayName := formValueToPointer(c, "display_name")
	description := formValueToPointer(c, "description")

	_, err := deps.Clients.Users.UpdateProfile(ctx, &proto.UpdateProfileRequest{Session: session, DisplayName: displayName, Description: description})
	if errors.Is(err, common.ErrSessionNotFound) {
		return c.Status(401).SendString("Please sign in to continue")
	} else if err != nil {
		return c.Status(500).SendString("There was an internal error: " + status.Convert(err).Message())
	}
	return Render(c, frontend.SoftSuccess("Saved"))
}
