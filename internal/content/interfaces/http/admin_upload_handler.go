package contenthttp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/son-ngo/edu-app/internal/content/application"
	shared "github.com/son-ngo/edu-app/internal/shared/domain"
	"github.com/son-ngo/edu-app/internal/shared/httpx"
	"github.com/son-ngo/edu-app/internal/shared/middleware"
)

// AdminUploadHandler exposes the admin upload + parse endpoints (ADMIN only).
type AdminUploadHandler struct {
	svc      *application.AdminUploadService
	validate middleware.TokenValidator
}

// NewAdminUploadHandler builds the handler.
func NewAdminUploadHandler(svc *application.AdminUploadService, validate middleware.TokenValidator) *AdminUploadHandler {
	return &AdminUploadHandler{svc: svc, validate: validate}
}

// Routes mounts the endpoints under /admin/uploads (all ADMIN-guarded).
func (h *AdminUploadHandler) Routes(rg *gin.RouterGroup) {
	g := rg.Group("/admin/uploads", middleware.Auth(h.validate), middleware.RequireRole(middleware.RoleAdmin))
	g.POST("/init", h.initUpload)
	g.POST("/complete", h.completeUpload)
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.POST("/:id/parse", h.retryParse)
	g.GET("/:id/parse-jobs", h.listParseJobs)
	g.DELETE("/:id", h.delete)
}

type initRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
}

func (h *AdminUploadHandler) initUpload(c *gin.Context) {
	var req initRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	res, err := h.svc.InitUpload(c.Request.Context(), application.InitInput{
		UploadedBy: middleware.UserIDFrom(c), Filename: req.Filename,
		ContentType: req.ContentType, FileSize: req.FileSize,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, gin.H{
		"asset_id":   res.Asset.ID,
		"object_key": res.Asset.ObjectKey,
		"upload_url": res.Upload.URL,
		"method":     res.Upload.Method,
		"headers":    res.Upload.Headers,
		"expires_at": res.Upload.ExpiresAt,
	})
}

type completeRequest struct {
	AssetID string `json:"asset_id" binding:"required"`
}

func (h *AdminUploadHandler) completeUpload(c *gin.Context) {
	var req completeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, shared.ErrValidation.WithCause(err))
		return
	}
	res, err := h.svc.CompleteUpload(c.Request.Context(), middleware.UserIDFrom(c), req.AssetID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"asset": res.Asset, "parse_job_id": res.ParseJobID})
}

func (h *AdminUploadHandler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	assets, total, err := h.svc.ListAssets(c.Request.Context(), c.Query("status"), limit, offset)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.List(c, assets, httpx.Meta{Total: total, Page: offset/maxInt(limit, 1) + 1, Limit: limit})
}

func (h *AdminUploadHandler) get(c *gin.Context) {
	asset, err := h.svc.GetAsset(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, asset)
}

func (h *AdminUploadHandler) retryParse(c *gin.Context) {
	job, err := h.svc.RetryParse(c.Request.Context(), middleware.UserIDFrom(c), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, job)
}

func (h *AdminUploadHandler) listParseJobs(c *gin.Context) {
	jobs, err := h.svc.ListParseJobs(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, jobs)
}

func (h *AdminUploadHandler) delete(c *gin.Context) {
	if err := h.svc.DeleteAsset(c.Request.Context(), c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"message": "asset deleted"})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
