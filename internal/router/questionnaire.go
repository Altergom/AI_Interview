package router

import (
	"github.com/cloudwego/hertz/pkg/route"

	"ai_interview/internal/middleware/ratelimit"
)

func (r *Router) registerQuestionnaire(v1 *route.RouterGroup) {
	questionnaire := v1.Group("/questionnaire", r.hauth)
	questionnaire.GET("", r.questionnaire.Get)
	questionnaire.POST("/submit",
		ratelimit.Middleware(r.rdb, "questionnaire.submit"),
		r.questionnaire.Submit,
	)
}

