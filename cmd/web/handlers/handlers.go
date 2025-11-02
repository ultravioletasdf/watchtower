package handlers

import (
	"context"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"

	"videoapp/cmd/web/frontend"
	"videoapp/internal/generated/proto"
	"videoapp/internal/utils"
)

type Dependencies struct {
	Config  utils.Config
	Clients struct {
		Users      proto.UsersClient
		Sessions   proto.SessionsClient
		Videos     proto.VideosClient
		Thumbnails proto.ThumbnailsClient
		Reactions  proto.ReactionsClient
	}
}

var deps Dependencies
var ctx = context.Background()

func Add(app *fiber.App, dependencies Dependencies) {
	deps = dependencies
	frontend.Init(deps.Config)
	app.Get("/", root)

	app.Get("/recommendations", getRecommendations)

	app.Get("/sign/in", signIn)
	app.Post("/sign/in", postSignIn)
	app.Get("/sign/up", signUp)
	app.Post("/sign/up", postSignUp)
	app.Get("/sign/out", signOut)

	app.Get("/settings", settings)
	app.Put("/settings/profile", putProfile)
	app.Post("/avatar", uploadAvatar)
	app.Delete("/avatar", deleteAvatar)

	app.Get("/user/:username", profile)
	app.Get("/user/id/:id", profileFromId)
	app.Get("/user/:id/:type", getFollowsModal)
	app.Get("/user/:id/extrainfo", extraUserInfo)
	app.Post("/stages", getStages)

	app.Post("/follow/:id", follow)
	app.Delete("/follow/:id", follow)

	app.Get("/videos/:id", viewVideo)
	app.Get("/videos/:id/edit", editVideo)
	app.Put("/videos/:id", saveVideoChanges)
	app.Put("/reactions/:id/:type", react)
	app.Delete("/reactions/:id", deleteReaction)
	app.Post("/videos/:id/comments", createComment)
	app.Get("/videos/:id/comments", listComments)
	app.Get("/status/:id", videoStatus)

	app.Get("/following", following)

	app.Get("/upload", upload)
	app.Post("/upload/video", uploadVideo)
	app.Get("/upload/:id", afterUpload)
	app.Post("/upload/:id/publish", publishVideo)
	app.Post("/upload/thumbnail", uploadThumbnail)
	app.Post("/upload/thumbnail/:id/process", processThumbnail)
	app.Post("/toasts", func(c *fiber.Ctx) error {
		return Render(c, frontend.SoftError("This is a toast"))
	})
	app.Use(etag.New())
	app.Static("/assets", "./cmd/web/assets", fiber.Static{Compress: true})
}
func root(c *fiber.Ctx) error {
	return Render(c, frontend.Home(getUser(c)))
}
func Render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c.Context(), c.Response().BodyWriter())
}
