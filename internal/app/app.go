package app

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/CosmoS1X/go-project-278/internal/config"
	"github.com/CosmoS1X/go-project-278/internal/service/links"
	"github.com/CosmoS1X/go-project-278/internal/service/visits"
	"github.com/CosmoS1X/go-project-278/internal/storage/sqlc"
)

func NewRouter(db sqlc.DBTX, cfg *config.Config) *gin.Engine {
	queries := sqlc.New(db)
	repo := links.NewRepository(queries)
	visitsRepo := visits.NewRepository(queries)
	handler := links.NewHandler(repo, visitsRepo, cfg.BaseShortURL)
	visitsHandler := visits.NewHandler(visitsRepo)

	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{cfg.CORSOrigin},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:  []string{"Content-Type"},
		ExposeHeaders: []string{"Content-Range"},
	}))
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/r/:code", handler.Redirect)

	api := router.Group("/api/links")
	api.GET("", handler.List)
	api.POST("", handler.Create)
	api.GET("/:id", handler.Get)
	api.PUT("/:id", handler.Update)
	api.DELETE("/:id", handler.Delete)

	router.GET("/api/link_visits", visitsHandler.ListVisits)

	return router
}
