package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/codepnw/go-starter-kit/internal/config"
	acchandler "github.com/codepnw/go-starter-kit/internal/features/account/handler"
	accrepository "github.com/codepnw/go-starter-kit/internal/features/account/repository"
	accservice "github.com/codepnw/go-starter-kit/internal/features/account/service"
	userhandler "github.com/codepnw/go-starter-kit/internal/features/user/handler"
	userrepository "github.com/codepnw/go-starter-kit/internal/features/user/repository"
	userservice "github.com/codepnw/go-starter-kit/internal/features/user/service"
	"github.com/codepnw/go-starter-kit/internal/middleware"
	"github.com/codepnw/go-starter-kit/pkg/database"
	jwttoken "github.com/codepnw/go-starter-kit/pkg/jwttoken"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	db     *sql.DB
	router *gin.Engine
	token  jwttoken.JWTToken
	mid    *middleware.Middleware
	tx     database.TxManager

	// Handler
	handlerUser    *userhandler.UserHandler
	handlerAccount *acchandler.AccountHandler
}

func NewServer(cfg *config.EnvConfig, db *sql.DB) (*Server, error) {
	r := gin.New()

	// JWT Token
	token, err := jwttoken.NewJWTToken(cfg.JWT.AppName, cfg.JWT.SecretKey, cfg.JWT.RefreshKey)
	if err != nil {
		return nil, err
	}

	// Middleware
	mid := middleware.InitMiddleware(token)

	// DB Transaction
	tx := database.NewDBTransaction(db)

	// Denpendency Injection
	s := &Server{
		db:     db,
		router: r,
		token:  token,
		mid:    mid,
		tx:     tx,
	}

	// Gin Middleware
	s.setupGinMiddleware()

	// Setup Routes Handler
	s.setupRoutesHandler()

	// Prefix Default: /api/v1
	prefix := s.router.Group(cfg.APP.Prefix)

	// Register Routes
	s.registerHealthRoutes(prefix)
	s.registerUserRoutes(prefix)
	s.registerAccountRoutes(prefix)

	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) setupGinMiddleware() {
	s.router.Use(gin.Recovery())
	s.router.Use(s.mid.Logger())
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func (s *Server) setupRoutesHandler() {
	// User Setup
	userRepo := userrepository.NewUserRepository(s.db)
	userSrv := userservice.NewUserService(s.tx, s.token, userRepo)
	s.handlerUser = userhandler.NewUserHandler(userSrv)

	// Account Setup
	accRepo := accrepository.NewAccountRepository(s.db)
	accSrv := accservice.NewAccountSevice(accRepo)
	s.handlerAccount = acchandler.NewAccountHandler(accSrv)
}
