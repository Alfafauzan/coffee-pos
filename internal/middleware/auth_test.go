package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func setupTestRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/protected", AuthMiddleware(secret), func(c *gin.Context) {
		userID := c.GetString("user_id")
		role := c.GetString("role")
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	return r
}

func generateTestToken(secret, userID, role string, expireDuration time.Duration) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(expireDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := setupTestRouter("my-test-secret")

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Token tidak ditemukan" {
		t.Errorf("expected 'Token tidak ditemukan', got: %v", res["message"])
	}
}

func TestAuthMiddleware_InvalidPrefix(t *testing.T) {
	r := setupTestRouter("my-test-secret")

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Format token tidak valid" {
		t.Errorf("expected 'Format token tidak valid', got: %v", res["message"])
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "my-test-secret"
	r := setupTestRouter(secret)

	// Token expired 1 hour ago
	token := generateTestToken(secret, "user-123", "cashier", -1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Token tidak valid atau sudah expired" {
		t.Errorf("expected 'Token tidak valid atau sudah expired', got: %v", res["message"])
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	r := setupTestRouter("my-test-secret")

	// Signed with different secret
	token := generateTestToken("different-secret", "user-123", "cashier", 1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Token tidak valid atau sudah expired" {
		t.Errorf("expected 'Token tidak valid atau sudah expired', got: %v", res["message"])
	}
}

func TestAuthMiddleware_InvalidSigningMethod(t *testing.T) {
	r := setupTestRouter("my-test-secret")

	// None signing method
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": "user-123",
		"role":    "cashier",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingClaims(t *testing.T) {
	secret := "my-test-secret"
	r := setupTestRouter(secret)

	// Token without user_id claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secret))

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Token tidak valid" {
		t.Errorf("expected 'Token tidak valid', got: %v", res["message"])
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "my-test-secret"
	r := setupTestRouter(secret)

	token := generateTestToken(secret, "user-abc-123", "owner", 1*time.Hour)

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["user_id"] != "user-abc-123" {
		t.Errorf("expected user_id user-abc-123, got: %v", res["user_id"])
	}
	if res["role"] != "owner" {
		t.Errorf("expected role owner, got: %v", res["role"])
	}
}

func setupRoleTestRouter(allowedRoles ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/role-protected", RoleMiddleware(allowedRoles...), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func TestRoleMiddleware_EmptyRole(t *testing.T) {
	r := setupRoleTestRouter("owner")

	req, _ := http.NewRequest(http.MethodGet, "/role-protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Token tidak valid" {
		t.Errorf("expected 'Token tidak valid', got: %v", res["message"])
	}
}

func TestRoleMiddleware_Forbidden(t *testing.T) {
	r := gin.New()
	r.GET("/role-protected", func(c *gin.Context) {
		c.Set("role", "cashier")
		c.Next()
	}, RoleMiddleware("owner"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/role-protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "Akses ditolak" {
		t.Errorf("expected 'Akses ditolak', got: %v", res["message"])
	}
}

func TestRoleMiddleware_Success(t *testing.T) {
	r := gin.New()
	r.GET("/role-protected", func(c *gin.Context) {
		c.Set("role", "cashier")
		c.Next()
	}, RoleMiddleware("owner", "cashier"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/role-protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
