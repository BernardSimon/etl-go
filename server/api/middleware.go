package api

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"strings"

	_type "github.com/BernardSimon/etl-go/server/type"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetRealIP(c *gin.Context) string {
	ip := c.Request.Header.Get("X-Forwarded-For")
	if ip != "" {
		if idx := strings.Index(ip, ","); idx != -1 {
			ip = ip[:idx]
		}
	}
	if ip == "" {
		ip = c.Request.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = c.ClientIP()
	}
	return ip
}

func RequestResponseMiddleware(c *gin.Context) {
	var log _type.RequestLog
	log.Method = c.Request.Method
	log.Ip = GetRealIP(c)
	log.Path = c.Request.URL.Path
	// 填充请求头
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if key == "Authorization" {
			continue
		}
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	log.Headers = headers
	token := c.Request.Header.Get("Authorization")
	c.Set("token", token)
	language := c.Request.Header.Get("Accept-Language")
	langMatched := false
	for _, lang := range i18n.AcceptLanguages {
		if lang == language {
			langMatched = true
			break
		}
	}
	if !langMatched {
		language = "en"
	}
	c.Set("language", language)
	requestBody, err := io.ReadAll(c.Request.Body)
	_ = c.Request.Body.Close()
	if err != nil {
		log.Body = "Fail To Get Body"
		c.Set("code", 1)
		c.Set("message", "Fail To Get Body")
		c.Abort()
	} else {
		stringBody := string(requestBody)
		log.Body = stringBody
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	// 执行后续处理
	c.Next()
	code := c.GetInt("code")
	message := c.GetString("message")
	maskData := c.GetString("maskData")
	if maskData != "" {
		log.Body = "mask"
	}
	response := _type.ResponseModel{
		Code:    code,
		Message: i18n.Translate(language, message),
	}
	log.Response = &response
	switch code {
	case 0:
		data, exists := c.Get("data")
		if exists {
			responseData := _type.ResponseWithData{
				Code:    code,
				Message: i18n.Translate(language, message),
				Data:    data,
			}
			c.JSON(200, responseData)
		} else {
			c.JSON(200, response)
		}
		zap.L().Info("request success", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
	case 1:
		zap.L().Warn("request public error", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
		c.JSON(400, response)
	case 2:
		zap.L().Warn("request service error", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
		c.JSON(400, response)
	case 3:
		zap.L().Warn("request auth error", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
		c.JSON(400, response)
	default:
		zap.L().Error("unknown request error", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
		c.JSON(500, response)
	}
	return
}

func Md5(str string) string {
	hash := md5.New()
	_, _ = io.WriteString(hash, str)
	return hex.EncodeToString(hash.Sum(nil))
}
