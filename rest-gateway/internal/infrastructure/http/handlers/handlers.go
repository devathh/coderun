package handlers

import (
	"fmt"
	"net/http"

	"github.com/devathh/coderun/rest-gateway/internal/application/services"
	"github.com/devathh/coderun/rest-gateway/internal/infrastructure/config"
	customerrors "github.com/devathh/coderun/rest-gateway/pkg/errors"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config, service services.RestGatewayService) (http.Handler, error) {
	if cfg == nil {
		return nil, customerrors.ErrNilArgs
	}

	var router *gin.Engine

	switch cfg.App.Env {
	case "local", "dev":
		router = gin.Default()
	case "prod":
		router = gin.New()
		router.Use(gin.Recovery())
	default:
		return nil, fmt.Errorf("invalid environment")
	}

	routes := NewRoutes(service)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/register", routes.Register())
			v1.POST("/login", routes.Login())
			v1.POST("/refresh", routes.Refresh())

			v1.PATCH("/user", routes.UpdateUser())
			v1.GET("/user", routes.GetSelf())
			v1.GET("/user/:id", routes.GetUserByID())
		}
	}

	return router, nil
}
