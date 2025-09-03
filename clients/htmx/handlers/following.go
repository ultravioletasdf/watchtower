package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"videoapp/clients/htmx/frontend"
	"videoapp/proto"
)

func following(c *fiber.Ctx) error {
	u := getUser(c)
	if u == nil {
		return c.Redirect("/sign/in")
	}

	page, _ := strconv.Atoi(c.Query("page"))

	videos, err := deps.Clients.Users.GetFollowingVideos(ctx, &proto.GetFollowingVideosRequest{UserId: u.Id, Page: int32(page)})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Send([]byte(err.Error()))
	}

	if page > 0 {
		return Render(c, frontend.FollowingVideos(videos.Videos, page+1))
	}

	return Render(c, frontend.Following(u, videos.Videos))
}
