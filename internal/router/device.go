package router

import "github.com/cloudwego/hertz/pkg/route"

func (r *Router) registerDevice(v1 *route.RouterGroup) {
	v1.POST("/device/check", r.device.Check)
}

