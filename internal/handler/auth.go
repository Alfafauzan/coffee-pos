package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Alfafauzan/coffee-pos/internal/entity"
	"github.com/Alfafauzan/coffee-pos/internal/service"
	"github.com/Alfafauzan/coffee-pos/pkg/response"
	"github.com/Alfafauzan/coffee-pos/pkg/validator"
)

// AuthHandler handles HTTP requests for user authentication and registration.
type AuthHandler struct {
	authService *service.AuthService
	validator   *validator.Validator
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService *service.AuthService, v *validator.Validator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   v,
	}
}

// Register processes customer/staff user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req entity.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		response.ValidationError(c, errs)
		return
	}

	result, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Registrasi berhasil", result)
}

// Login processes user authentication and generates a JWT access token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Format request tidak valid")
		return
	}

	if errs := h.validator.Validate(req); len(errs) > 0 {
		response.ValidationError(c, errs)
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Login berhasil", result)
}
