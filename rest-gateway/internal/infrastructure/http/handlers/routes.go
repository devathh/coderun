package handlers

import (
	"net/http"

	"github.com/devathh/coderun/rest-gateway/internal/application/dto"
	"github.com/devathh/coderun/rest-gateway/internal/application/services"
	"github.com/gin-gonic/gin"
)

type Routes struct {
	service services.RestGatewayService
}

func NewRoutes(service services.RestGatewayService) *Routes {
	return &Routes{
		service: service,
	}
}

func (r *Routes) Register() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.RegisterRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Register(ctx, &req)
		if err != nil {
			ctx.AbortWithStatusJSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.LoginRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Login(ctx, &req)
		if err != nil {
			ctx.AbortWithStatusJSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) Refresh() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req dto.RefreshRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		resp, code, err := r.service.Refresh(ctx, &req)
		if err != nil {
			ctx.AbortWithStatusJSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) UpdateUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("session")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		var req dto.UpdateRequest
		if err := ctx.BindJSON(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}

		code, err := r.service.UpdateUser(ctx, &req, token)
		if err != nil {
			ctx.JSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, nil)
	}
}

func (r *Routes) GetUserByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.Param("id")

		resp, code, err := r.service.GetUserByID(ctx, &dto.GetByIDRequest{
			UserID: userID,
		})
		if err != nil {
			ctx.AbortWithStatusJSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}

func (r *Routes) GetSelf() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie("session")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		resp, code, err := r.service.GetSelf(ctx, token)
		if err != nil {
			ctx.AbortWithStatusJSON(code, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(code, resp)
	}
}
