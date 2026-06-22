package router

import (
	"github.com/cloudwego/hertz/pkg/route"

	"ai_interview/internal/middleware/ratelimit"
)

func (r *Router) registerResume(v1 *route.RouterGroup) {
	resume := v1.Group("/resume", r.hauth)
	resume.GET("/upload-url", r.resume.PresignUpload)
	resume.POST("/parse", ratelimit.Middleware(r.rdb, "resume.parse"), r.resume.Parse)
	resume.POST("/submit", r.resume.Submit)
	resume.GET("", r.resume.Get)
}

