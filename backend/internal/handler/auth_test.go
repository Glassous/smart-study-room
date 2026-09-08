package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

func TestLogoutReturnsServiceUnavailableWithoutRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(service.NewAuthService(nil, &config.Config{JWTSecret: "test"}, nil))
	r := gin.New()
	r.POST("/logout", h.Logout)
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
