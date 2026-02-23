package server

import (
	"net/http"

	"github.com/codepnw/go-starter-kit/pkg/utils/response"
	"github.com/gin-gonic/gin"
)

func (s *Server) registerHealthRoutes(r *gin.RouterGroup) {
	r.GET("/health", func(c *gin.Context) {
		response.ResponseMessage(c, http.StatusOK, "Core Bank API Running...")
	})
}

func (s *Server) registerUserRoutes(r *gin.RouterGroup) {
	handler := s.handlerUser

	// Auth Routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)

		// Authorized
		auth.POST("/logout", handler.Logout, s.mid.Authorized())
	}

	// Users Routes
	users := r.Group("/users", s.mid.Authorized())
	{
		users.GET("/profile", handler.GetProfile)
	}
}

func (s *Server) registerAccountRoutes(r *gin.RouterGroup) {
	handler := s.handlerAccount

	accounts := r.Group("/accounts", s.mid.Authorized())
	{
		accounts.POST("/", handler.CreateAccount)
		accounts.GET("/", handler.GetAllAccounts)
		accounts.GET("/:id", handler.GetAccountByID)
		accounts.POST("/:id/deposit", handler.DepositMoney)
	}
}
