package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"

	"ai_interview/internal/domain"
	authmw "ai_interview/internal/middleware/auth"
	"ai_interview/internal/service"
	biz "ai_interview/internal/utils/respx"
)

type resumeHandler struct {
	svc service.ResumeService
}

// PresignUpload GET /v1/resume/upload-url?filename=xxx.pdf
// 生成前端直传 S3 的预签名 PUT URL（5 分钟有效）。
// 响应: { upload_url, object_key }
func (h *resumeHandler) PresignUpload(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	filename := string(c.Query("filename"))
	if filename == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "filename is required")
		return
	}

	uploadURL, objectKey, err := h.svc.PresignUpload(ctx, userID, filename)
	if err != nil {
		Fail(ctx, c, http.StatusInternalServerError, biz.CodeStorageUploadFailed, "generate upload url failed")
		return
	}

	OK(ctx, c, map[string]string{
		"upload_url": uploadURL,
		"object_key": objectKey,
	})
}

// parseReq POST /v1/resume/parse 请求体。
type parseReq struct {
	ObjectKey string `json:"object_key"`
}

// Parse POST /v1/resume/parse
func (h *resumeHandler) Parse(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	var req parseReq
	if err := c.BindJSON(&req); err != nil || req.ObjectKey == "" {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "object_key is required")
		return
	}

	resume, err := h.svc.Parse(ctx, userID, req.ObjectKey)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, resume)
}

// submitProject handler 层简历项目 DTO，与 domain.ResumeProject 字段对齐。
type submitProject struct {
	Name        string   `json:"name"`
	TechStack   []string `json:"tech_stack"`
	Description string   `json:"description"`
	Highlights  []string `json:"highlights"`
}

// submitInternship handler 层实习经历 DTO。
type submitInternship struct {
	Company     string `json:"company,omitempty"`
	Role        string `json:"role,omitempty"`
	Description string `json:"description,omitempty"`
}

// submitEducation handler 层教育背景 DTO。
type submitEducation struct {
	School     string `json:"school"`
	Major      string `json:"major"`
	Graduation string `json:"graduation"`
}

// submitReq POST /v1/resume/submit 请求体，使用独立 DTO 不引用 domain 类型。
type submitReq struct {
	Name        string             `json:"name,omitempty"`
	Phone       string             `json:"phone,omitempty"`
	Email       string             `json:"email,omitempty"`
	Skills      []string           `json:"skills"`
	Projects    []submitProject    `json:"projects"`
	Internships []submitInternship `json:"internships"`
	Education   submitEducation    `json:"education"`
}

// Submit POST /v1/resume/submit
func (h *resumeHandler) Submit(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	var req submitReq
	if err := c.BindJSON(&req); err != nil {
		Fail(ctx, c, http.StatusBadRequest, biz.CodeBadRequest, "invalid request body")
		return
	}

	projects := make([]domain.ResumeProject, len(req.Projects))
	for i, p := range req.Projects {
		projects[i] = domain.ResumeProject{
			Name:        p.Name,
			TechStack:   p.TechStack,
			Description: p.Description,
			Highlights:  p.Highlights,
		}
	}
	internships := make([]domain.ResumeInternship, len(req.Internships))
	for i, n := range req.Internships {
		internships[i] = domain.ResumeInternship{
			Company:     n.Company,
			Role:        n.Role,
			Description: n.Description,
		}
	}

	resume := domain.StructuredResume{
		UserID:      userID,
		Name:        req.Name,
		Phone:       req.Phone,
		Email:       req.Email,
		Skills:      req.Skills,
		Projects:    projects,
		Internships: internships,
		Education: domain.ResumeEducation{
			School:     req.Education.School,
			Major:      req.Education.Major,
			Graduation: req.Education.Graduation,
		},
	}

	resumeID, err := h.svc.Submit(ctx, resume)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}

	OK(ctx, c, map[string]string{"resume_id": resumeID})
}

// Get GET /v1/resume
func (h *resumeHandler) Get(ctx context.Context, c *app.RequestContext) {
	userID := authmw.GetUserID(c)

	resume, err := h.svc.Get(ctx, userID)
	if err != nil {
		HandleErr(ctx, c, err)
		return
	}
	OK(ctx, c, resume)
}
