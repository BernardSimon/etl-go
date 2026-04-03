package api

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/BernardSimon/etl-go/server/config"
	_type "github.com/BernardSimon/etl-go/server/type"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

// loginRateLimiter implements per-IP rate limiting for login attempts.
var loginLimiters = struct {
	sync.RWMutex
	m map[string]*rate.Limiter
}{m: make(map[string]*rate.Limiter)}

func getLoginLimiter(ip string) *rate.Limiter {
	loginLimiters.RLock()
	limiter, exists := loginLimiters.m[ip]
	loginLimiters.RUnlock()
	if exists {
		return limiter
	}
	// Allow 5 login attempts per minute per IP
	limiter = rate.NewLimiter(rate.Every(time.Minute/5), 5)
	loginLimiters.Lock()
	loginLimiters.m[ip] = limiter
	loginLimiters.Unlock()
	return limiter
}

func Login(req *_type.LoginRequest, _ string) (interface{}, error) {
	if req.Username == config.Config.Username && req.Password == config.Config.Password {
		token, err := generateToken(req.Username)
		if err != nil {
			return nil, errors.New("failed to generate token")
		}
		response := _type.LoginResponse{
			Token: token,
		}
		return response, nil
	}
	return nil, errors.New("invalid username or password")
}

// LoginWithRateLimit wraps Login with per-IP rate limiting.
func LoginWithRateLimit(c *gin.Context) func(*_type.LoginRequest, string) (interface{}, error) {
	ip := GetRealIP(c)
	limiter := getLoginLimiter(ip)
	return func(req *_type.LoginRequest, lang string) (interface{}, error) {
		if !limiter.Allow() {
			return nil, errors.New("too many login attempts, please try again later")
		}
		return Login(req, lang)
	}
}

func generateToken(UserId string) (string, error) {
	// 创建声明
	notBefore := jwt.NewNumericDate(time.Now())
	claims := &jwt.RegisteredClaims{
		Subject:   UserId,
		NotBefore: notBefore,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 6)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Config.JwtSecret))
}

func DecodeToken(tokenString string) (string, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.JwtSecret), nil
	})
	if err != nil {
		return "", err
	}
	if token == nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	userId := claims.Subject
	return userId, nil
}

func AuthMiddleware(c *gin.Context) {
	token := c.GetString("token")
	// Support "Bearer <token>" format
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	_, err := DecodeToken(token)
	if err != nil {
		c.Set("code", 3)
		c.Set("message", "invalid token")
		c.Abort()
		return
	}
	c.Next()
}

func AuthMiddlewareFile(c *gin.Context) {
	token := c.Query("token")
	_, err := DecodeToken(token)
	if err != nil {
		c.JSON(400, gin.H{
			"code":    3,
			"message": "invalid token",
		})
		c.Abort()
		return
	}
	c.Next()
}
