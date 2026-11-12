package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Alfafauzan/coffee-pos/pkg/response"
)

// AuthMiddleware validates JWT Bearer tokens from the Authorization header
// and populates user_id and role in the Gin context.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil header Authorization dari request
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Token tidak ditemukan")
			c.Abort()
			return
		}

		// 2. Cek apakah header dimulai dengan "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, "Format token tidak valid")
			c.Abort()
			return
		}

		// 3. Potong prefix "Bearer " untuk dapat token string-nya
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 4. Parse dan validasi token dengan jwt.Parse()
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("signing method tidak valid")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			response.Unauthorized(c, "Token tidak valid atau sudah expired")
			c.Abort()
			return
		}

		// 5. Ekstrak claims dari token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "Token tidak valid")
			c.Abort()
			return
		}

		// 6. Ambil user_id dan role dari claims
		userID, okUserID := claims["user_id"].(string)
		role, okRole := claims["role"].(string)
		if !okUserID || !okRole {
			response.Unauthorized(c, "Token tidak valid")
			c.Abort()
			return
		}

		// 7. Simpan ke Gin context
		c.Set("user_id", userID)
		c.Set("role", role)

		// 8. c.Next() — teruskan ke handler berikutnya
		c.Next()
	}
}

// RoleMiddleware checks if the authenticated user's role matches any of the allowed roles.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil role dari Gin context
		role := c.GetString("role")
		if role == "" {
			response.Unauthorized(c, "Token tidak valid")
			c.Abort()
			return
		}

		// 2. Cek apakah role ada di dalam allowedRoles
		allowed := false
		for _, r := range allowedRoles {
			if r == role {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "Akses ditolak")
			c.Abort()
			return
		}

		// 3. c.Next()
		c.Next()
	}
}

