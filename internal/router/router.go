package router

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"

	authmw "ai_interview/internal/middleware/auth"
)

type Auth interface {
	Register(ctx context.Context, c *app.RequestContext)
	Login(ctx context.Context, c *app.RequestContext)
	Guest(ctx context.Context, c *app.RequestContext)
}

type Device interface {
	Check(ctx context.Context, c *app.RequestContext)
}

type Resume interface {
	PresignUpload(ctx context.Context, c *app.RequestContext)
	Parse(ctx context.Context, c *app.RequestContext)
	Submit(ctx context.Context, c *app.RequestContext)
	Get(ctx context.Context, c *app.RequestContext)
}

type Interview interface {
	Config(ctx context.Context, c *app.RequestContext)
	Create(ctx context.Context, c *app.RequestContext)
	Stream(ctx context.Context, c *app.RequestContext)
	Audio(ctx context.Context, c *app.RequestContext)
	Finish(ctx context.Context, c *app.RequestContext)
	State(ctx context.Context, c *app.RequestContext)
	CodeSubmit(ctx context.Context, c *app.RequestContext)
}

type WSInterview interface {
	ServeWS(ctx context.Context, c *app.RequestContext)
}

type Report interface {
	Status(ctx context.Context, c *app.RequestContext)
	Get(ctx context.Context, c *app.RequestContext)
}

type Questionnaire interface {
	Get(ctx context.Context, c *app.RequestContext)
	Submit(ctx context.Context, c *app.RequestContext)
}

type Router struct {
	rdb   *redis.Client
	hauth app.HandlerFunc

	auth          Auth
	device        Device
	resume        Resume
	interview     Interview
	wsInterview   WSInterview
	report        Report
	questionnaire Questionnaire
}

type Deps struct {
	JWTSecret string
	Rdb       *redis.Client

	Auth          Auth
	Device        Device
	Resume        Resume
	Interview     Interview
	WSInterview   WSInterview
	Report        Report
	Questionnaire Questionnaire
}

func Register(h *server.Hertz, deps Deps) {
	r := &Router{
		rdb:           deps.Rdb,
		hauth:         authmw.HAuth(deps.JWTSecret),
		auth:          deps.Auth,
		device:        deps.Device,
		resume:        deps.Resume,
		interview:     deps.Interview,
		wsInterview:   deps.WSInterview,
		report:        deps.Report,
		questionnaire: deps.Questionnaire,
	}
	r.register(h)
}

func (r *Router) register(h *server.Hertz) {
	r.registerPublic(h)

	v1 := h.Group("/v1")
	r.registerAuth(v1)
	r.registerDevice(v1)
	r.registerResume(v1)
	r.registerInterview(v1)
	r.registerWS(v1)
	r.registerReport(v1)
	r.registerQuestionnaire(v1)
}
