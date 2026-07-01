package router

import "github.com/cloudwego/hertz/pkg/route"

func (r *Router) registerReport(v1 *route.RouterGroup) {
	report := v1.Group("/report", r.hauth)
	report.GET("/status", r.report.Status)
	report.GET("", r.report.Get)
}

