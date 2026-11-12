package handler

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"github.com/Alfafauzan/coffee-pos/config"
	"github.com/Alfafauzan/coffee-pos/internal/entity"
	"github.com/Alfafauzan/coffee-pos/internal/middleware"
	"github.com/Alfafauzan/coffee-pos/internal/repository"
	"github.com/Alfafauzan/coffee-pos/internal/service"
	"github.com/Alfafauzan/coffee-pos/pkg/response"
	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

// NewRouter initializes and configures the Gin engine router with dependencies.
func NewRouter(db *sql.DB, cfg *config.Config, v *validator.Validator) *gin.Engine {
	if cfg != nil && cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// 1. Inisialisasi repository
	userRepo := repository.NewUserRepository(db)

	// 2. Inisialisasi service
	var jwtSecret string
	var jwtExpireHours int
	if cfg != nil {
		jwtSecret = cfg.JWTSecret
		jwtExpireHours = cfg.JWTExpireHours
	}
	authService := service.NewAuthService(userRepo, jwtSecret, jwtExpireHours)

	// 3. Inisialisasi handler
	authHandler := NewAuthHandler(authService, v)

	// 4. Daftarkan route group /api/v1
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			response.OK(c, "server is running", nil)
		})

		// 5. Tambahkan route group /api/v1/auth (tanpa middleware auth)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Route group untuk owner — semua endpoint di sini butuh login + role owner
		ownerGroup := v1.Group("/owner")
		ownerGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		ownerGroup.Use(middleware.RoleMiddleware(entity.RoleOwner))
		_ = ownerGroup

		// Route group untuk cashier — semua endpoint di sini butuh login + role cashier
		cashierGroup := v1.Group("/cashier")
		cashierGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		cashierGroup.Use(middleware.RoleMiddleware(entity.RoleCashier))
		_ = cashierGroup
	}

	return router
}
