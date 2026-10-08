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
	linksRepo := links.NewRepository(queries)
	visitsRepo := visits.NewRepository(queries)
	linksHandler := links.NewHandler(linksRepo, visitsRepo, cfg.BaseShortURL)
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

	router.GET("/r/:code", linksHandler.Redirect)

	api := router.Group("/api/links")
	api.GET("", linksHandler.List)
	api.POST("", linksHandler.Create)
	api.GET("/:id", linksHandler.Get)
	api.PUT("/:id", linksHandler.Update)
	api.DELETE("/:id", linksHandler.Delete)

	router.GET("/api/link_visits", visitsHandler.List)

	return router
}
