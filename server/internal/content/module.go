package content

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/son-ngo/edu-app/internal/app"
	"github.com/son-ngo/edu-app/internal/content/application"
	"github.com/son-ngo/edu-app/internal/content/infrastructure"
	contenthttp "github.com/son-ngo/edu-app/internal/content/interfaces/http"
	"github.com/son-ngo/edu-app/internal/shared/cache"
	s3pkg "github.com/son-ngo/edu-app/pkg/s3"
)

const cacheTTL = 10 * time.Minute

func Register(rg *gin.RouterGroup, deps *app.Deps) {
	repo := infrastructure.NewPgRepository(deps.DB)
	svc := application.NewService(repo, application.WithCache(cache.NewRedisCache(deps.Redis), cacheTTL))
	contenthttp.NewHandler(svc, deps.AuthValidate).Routes(rg)

	registerUploads(rg, deps)
}

func registerUploads(rg *gin.RouterGroup, deps *app.Deps) {
	s3client, err := s3pkg.New(s3pkg.Config{
		Endpoint:     deps.Cfg.S3.Endpoint,
		Region:       deps.Cfg.S3.Region,
		AccessKey:    deps.Cfg.S3.AccessKey,
		SecretKey:    deps.Cfg.S3.SecretKey,
		Bucket:       deps.Cfg.S3.Bucket,
		UsePathStyle: deps.Cfg.S3.UsePathStyle,
		PresignTTL:   deps.Cfg.Upload.PresignTTL,
	})
	if err != nil {
		deps.Log.Warn("admin upload routes disabled: object storage not configured", zap.Error(err))
		return
	}

	storage := infrastructure.NewS3Storage(s3client)
	assets := infrastructure.NewPgAssetRepository(deps.DB)
	jobs := infrastructure.NewPgParseJobRepository(deps.DB)
	uploadSvc := application.NewAdminUploadService(assets, jobs, storage, deps.Cfg.Upload.MaxFileSizeBytes)

	contenthttp.NewAdminUploadHandler(uploadSvc, deps.AuthValidate).Routes(rg)
}
