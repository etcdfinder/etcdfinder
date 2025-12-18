package api

import (
	v1 "github.com/etcdfinder/etcdfinder/internal/api/v1"
	"github.com/etcdfinder/etcdfinder/internal/rest/middleware"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	EtcdFinderHandler *v1.EtcdfinderHandler
}

func NewRouter(handlers Handlers) (*gin.Engine, error) {
	// Set gin mode to release
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Disable trusted proxies as we are not running behind a proxy that sets forwarded headers
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	router.Use(
		middleware.LoggerMiddleware(),
		middleware.RequestIDMiddleware,
		middleware.CORSMiddleware,
		middleware.ErrorHandler(),
	)

	v1 := router.Group("/v1")

	{
		v1.POST("/get-key", handlers.EtcdFinderHandler.GetKey)
		v1.POST("/search-keys", handlers.EtcdFinderHandler.SearchKeys)
		v1.PUT("/put-key", handlers.EtcdFinderHandler.PutKey)
		v1.DELETE("/delete-key", handlers.EtcdFinderHandler.DeleteKey)
		v1.GET("/ingestion-delay", handlers.EtcdFinderHandler.GetIngestionDelay)
	}

	return router, nil
}
