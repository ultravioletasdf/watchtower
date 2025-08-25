package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"videoapp/clients/htmx/frontend"
	"videoapp/proto"
)

func viewVideo(c *fiber.Ctx) error {
	idInt, err := strconv.ParseInt(c.Params("id"), 10, 0)
	if err != nil {
		return c.SendString("Invalid video id")
	}
	v, err := deps.Clients.Videos.Get(ctx, &proto.GetVideoRequest{Session: c.Params("session"), Id: idInt})
	// add better error handling
	if shouldReturn := unwrapGrpcError(c, err, 500); shouldReturn {
		return nil
	}
	return Render(c, frontend.ViewVideo(getUser(c), v))
}
func videoStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 0)
	if err != nil {
		return c.SendString("Invalid id")
	}
	stage, err := deps.Clients.Videos.GetStage(ctx, &proto.VideosGetStageRequest{Session: c.Params("session"), Id: id})
	if shouldReturn := unwrapGrpcError(c, err, 200); shouldReturn {
		return nil
	}
	return Render(c, frontend.VideoStatusPoller(&proto.GetVideoResponse{Id: id, UploadId: stage.UploadId, Stage: stage.Stage}))
}
