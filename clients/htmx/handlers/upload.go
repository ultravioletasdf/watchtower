package handlers

import (
	"fmt"
	"strconv"
	"videoapp/clients/htmx/frontend"
	"videoapp/proto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"
)

func upload(c *fiber.Ctx) error {
	return Render(c, frontend.Upload())
}

// Endpoint to get a presigned post request
func uploadVideo(c *fiber.Ctx) error {
	upload, err := deps.Clients.Videos.CreateUpload(ctx, &proto.CreateUploadRequest{Session: c.Cookies("session")})
	if err != nil {
		status, grpcError := status.FromError(err)
		if grpcError {
			return c.SendString(status.Message())
		}
		return c.SendString(fmt.Sprintf("Unknown Error: %s", err.Error()))
	}

	// Convert ID to string because javascript can't handle big numbers
	return c.JSON(map[string]any{"id": strconv.FormatInt(upload.Id, 10), "form_data": upload.FormData, "url": upload.Url})
}

func afterUpload(c *fiber.Ctx) error {
	return Render(c, frontend.AfterUpload(c.Params("id")))
}
func uploadThumbnail(c *fiber.Ctx) error {
	upload, err := deps.Clients.Thumbnails.CreateUpload(ctx, &proto.CreateUploadRequest{Session: c.Cookies("session")})
	if err != nil {
		status, grpcError := status.FromError(err)
		if grpcError {
			return c.SendString(status.Message())
		}
		return c.SendString(fmt.Sprintf("Unknown Error: %s", err.Error()))
	}

	// Convert ID to string because javascript can't handle big numbers
	return c.JSON(map[string]any{"id": strconv.FormatInt(upload.Id, 10), "form_data": upload.FormData, "url": upload.Url})
}
func processThumbnail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.SendStatus(400)
	}
	idInt, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return c.SendStatus(400)
	}
	_, err = deps.Clients.Thumbnails.Process(ctx, &proto.ThumbnailsProcessRequest{Id: idInt, Session: c.Cookies("session")})
	if shouldReturn := unwrapGrpcError(c, err, 400); shouldReturn {
		return nil
	}
	return c.SendStatus(200)
}

func publishVideo(c *fiber.Ctx) error {
	var visibility proto.Visibility
	switch c.FormValue("visibility") {
	case "Public":
		visibility = proto.Visibility_Public
	case "Private":
		visibility = proto.Visibility_Private
	case "Unlisted":
		visibility = proto.Visibility_Unlisted
	}
	uploadId, err := strconv.ParseInt(c.Params("id"), 10, 0)
	if err != nil {
		return c.SendString("Invalid upload id, try uploading again from scratch")
	}
	thumbnailId, err := strconv.ParseInt(c.FormValue("thumbnailId"), 10, 0)
	if err != nil {
		return c.SendString("Invalid thumbnail id, try uploading a new thumbnail")
	}
	v, err := deps.Clients.Videos.Create(ctx, &proto.VideosCreateRequest{Session: c.Cookies("session"), UploadId: uploadId, ThumbnailId: thumbnailId, Title: c.FormValue("title"), Description: c.FormValue("description"), Visibility: visibility})
	if shouldReturn := unwrapGrpcError(c, err, 200); shouldReturn {
		return nil
	}
	c.Set("HX-Redirect", fmt.Sprintf("/videos/%d", v.Id))
	return nil
}
