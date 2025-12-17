package api

import (
	"errors"

	"time"

	"github.com/BernardSimon/etl-go/server/config"
	_type "github.com/BernardSimon/etl-go/server/type"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

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
