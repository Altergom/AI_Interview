package router

import "github.com/cloudwego/hertz/pkg/route"

func (r *Router) registerAuth(v1 *route.RouterGroup) {
	auth := v1.Group("/auth")
	auth.POST("/register", r.auth.Register)
	auth.POST("/login", r.auth.Login)
	auth.POST("/guest", r.auth.Guest)
}

