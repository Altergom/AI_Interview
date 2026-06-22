package router

import "github.com/cloudwego/hertz/pkg/route"

func (r *Router) registerWS(v1 *route.RouterGroup) {
	v1.GET("/interview/ws/:interview_id", r.wsInterview.ServeWS)
}

