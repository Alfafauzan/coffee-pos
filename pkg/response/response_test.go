package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func executeHandler(handlerFunc gin.HandlerFunc) (*httptest.ResponseRecorder, Response) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handlerFunc(c)

	var resp Response
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w, resp
}

func TestOK(t *testing.T) {
	w, resp := executeHandler(func(c *gin.Context) {
		OK(c, "success fetch", map[string]string{"key": "value"})
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Message != "success fetch" {
		t.Errorf("expected message 'success fetch', got '%s'", resp.Message)
	}
}

func TestCreated(t *testing.T) {
	w, resp := executeHandler(func(c *gin.Context) {
		Created(c, "created successfully", 123)
	})

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
}

func TestPaginated(t *testing.T) {
	meta := Meta{
		Page:       1,
		PerPage:    10,
		Total:      50,
		TotalPages: 5,
	}
	w, resp := executeHandler(func(c *gin.Context) {
		Paginated(c, "list loaded", []string{"a", "b"}, meta)
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if resp.Meta == nil || resp.Meta.Total != 50 {
		t.Errorf("expected meta total 50, got %v", resp.Meta)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(c *gin.Context)
		expectCode int
		expectMsg  string
	}{
		{"BadRequest", func(c *gin.Context) { BadRequest(c, "bad request") }, http.StatusBadRequest, "bad request"},
		{"Unauthorized", func(c *gin.Context) { Unauthorized(c, "unauthorized") }, http.StatusUnauthorized, "unauthorized"},
		{"Forbidden", func(c *gin.Context) { Forbidden(c, "forbidden") }, http.StatusForbidden, "forbidden"},
		{"NotFound", func(c *gin.Context) { NotFound(c, "not found") }, http.StatusNotFound, "not found"},
		{"InternalError", func(c *gin.Context) { InternalError(c, "server error") }, http.StatusInternalServerError, "server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, resp := executeHandler(tt.fn)
			if w.Code != tt.expectCode {
				t.Errorf("expected status %d, got %d", tt.expectCode, w.Code)
			}
			if resp.Success {
				t.Errorf("expected success false, got true")
			}
			if resp.Message != tt.expectMsg {
				t.Errorf("expected message '%s', got '%s'", tt.expectMsg, resp.Message)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	errDetails := []map[string]string{{"field": "name", "error": "required"}}
	w, resp := executeHandler(func(c *gin.Context) {
		ValidationError(c, errDetails)
	})

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", w.Code)
	}
	if resp.Success {
		t.Errorf("expected success false, got true")
	}
	if resp.Message != "Validasi gagal" {
		t.Errorf("expected message 'Validasi gagal', got '%s'", resp.Message)
	}
	if resp.Errors == nil {
		t.Errorf("expected errors not nil")
	}
}
