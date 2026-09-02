package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/CosmoS1X/go-project-278/internal/config"
	"github.com/CosmoS1X/go-project-278/internal/service/links"
	"github.com/CosmoS1X/go-project-278/internal/storage/sqlc"
)

func NewRouter(db sqlc.DBTX, cfg *config.Config) *gin.Engine {
	repo := links.NewRepository(sqlc.New(db))
	handler := links.NewHandler(repo, cfg.BaseShortURL)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	api := router.Group("/api/links")
	api.GET("", handler.List)
	api.POST("", handler.Create)
	api.GET("/:id", handler.Get)
	api.PUT("/:id", handler.Update)
	api.DELETE("/:id", handler.Delete)

	return router
}
