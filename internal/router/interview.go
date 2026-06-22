package router

import "github.com/cloudwego/hertz/pkg/route"

func (r *Router) registerInterview(v1 *route.RouterGroup) {
	interview := v1.Group("/interview", r.hauth)
	interview.POST("/config", r.interview.Config)
	interview.POST("/create", r.interview.Create)
	interview.GET("/stream", r.interview.Stream)
	interview.POST("/audio", r.interview.Audio)
	interview.POST("/finish", r.interview.Finish)
	interview.GET("/state", r.interview.State)
	interview.POST("/code/submit", r.interview.CodeSubmit)
}

