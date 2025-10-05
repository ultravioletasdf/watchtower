package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"videoapp/cmd/web/frontend"
	"videoapp/internal/generated/proto"
)

func viewVideo(c *fiber.Ctx) error {
	idInt, err := strconv.ParseInt(c.Params("id"), 10, 0)
	if err != nil {
		return c.SendString("Invalid video id")
	}
	v, err := deps.Clients.Videos.Get(ctx, &proto.GetVideoRequest{Session: c.Cookies("session"), Id: idInt})
	if err != nil {
		status := status.Convert(err)
		if status.Code() == codes.PermissionDenied || status.Code() == codes.Unauthenticated {
			return Render(c, frontend.VideoError(getUser(c), "This video is private"))
		}
		return Render(c, frontend.VideoError(getUser(c), "Unknown Error: "+status.Message()))
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
	return Render(c, frontend.VideoStatusPoller(&proto.GetVideoResponse{Id: id, UploadId: stage.UploadId, Stage: stage.Stage, AuthorizationPayload: stage.AuthorizationPayload, AuthorizationSignature: stage.AuthorizationSignature}))
}
