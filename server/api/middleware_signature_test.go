package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/BernardSimon/etl-go/server/config"
	"github.com/gin-gonic/gin"
)

func TestValidateSignatureSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalSecret := config.Config.ApiSecret
	config.Config.ApiSecret = "demo-secret"
	defer func() {
		config.Config.ApiSecret = originalSecret
	}()

	body := "{\n  \"name\": \"demo\",\n  \"enabled\": true\n}"
	query := url.Values{}
	query.Set("page_no", "1")
	query.Set("page_size", "10")
	query.Set("timestamp", "1712300000")
	query.Set("sign", buildSignature(query, body))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks?"+query.Encode(), strings.NewReader(body))
	c.Request = req
	c.Set("rawBody", body)

	originalNow := timeNow
	timeNow = func() time.Time {
		return time.Unix(1712300000, 0)
	}
	defer func() {
		timeNow = originalNow
	}()

	if err := ValidateSignature(c); err != nil {
		t.Fatalf("expected signature validation success, got error: %v", err)
	}
}

func TestValidateSignatureExpired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalSecret := config.Config.ApiSecret
	config.Config.ApiSecret = "demo-secret"
	defer func() {
		config.Config.ApiSecret = originalSecret
	}()

	body := `{"name":"demo"}`
	query := url.Values{}
	query.Set("timestamp", "1712300000")
	query.Set("sign", buildSignature(query, body))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks?"+query.Encode(), strings.NewReader(body))
	c.Request = req
	c.Set("rawBody", body)

	originalNow := timeNow
	timeNow = func() time.Time {
		return time.Unix(1712300062, 0)
	}
	defer func() {
		timeNow = originalNow
	}()

	err := ValidateSignature(c)
	if err == nil || err.Error() != "api signature expired" {
		t.Fatalf("expected expired signature error, got: %v", err)
	}
}

func TestValidateSignatureDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalSecret := config.Config.ApiSecret
	config.Config.ApiSecret = ""
	defer func() {
		config.Config.ApiSecret = originalSecret
	}()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/tasks?timestamp=1&sign=abc", nil)

	err := ValidateSignature(c)
	if err == nil || err.Error() != "api signature auth disabled" {
		t.Fatalf("expected disabled error, got: %v", err)
	}
}
