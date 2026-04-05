package api

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BernardSimon/etl-go/server/config"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var timeNow = time.Now

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
	var log types.RequestLog
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
	// 前缀匹配：zh-CN → zh, en-US → en
	if !langMatched {
		primary := strings.Split(language, "-")[0]
		for _, lang := range i18n.AcceptLanguages {
			if lang == primary {
				language = primary
				langMatched = true
				break
			}
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
		c.Set("rawBody", stringBody)
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
	response := types.ResponseModel{
		Code:    code,
		Message: i18n.Translate(language, message),
	}
	log.Response = &response
	data, exists := c.Get("data")
	switch code {
	case 0:
		if exists {
			responseData := types.ResponseWithData{
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
		if exists {
			c.JSON(400, types.ResponseWithData{
				Code:    code,
				Message: i18n.Translate(language, message),
				Data:    data,
			})
		} else {
			c.JSON(400, response)
		}
	case 2:
		zap.L().Warn("request service error", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
		if exists {
			c.JSON(422, types.ResponseWithData{
				Code:    code,
				Message: i18n.Translate(language, message),
				Data:    data,
			})
		} else {
			c.JSON(422, response)
		}
	case 3:
		zap.L().Warn("request auth error", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
		if exists {
			c.JSON(401, types.ResponseWithData{
				Code:    code,
				Message: i18n.Translate(language, message),
				Data:    data,
			})
		} else {
			c.JSON(401, response)
		}
	default:
		zap.L().Error("unknown request error", zap.String("service", "request_log"), zap.Any("content", log), zap.String("name", Md5(token)))
		if exists {
			c.JSON(500, types.ResponseWithData{
				Code:    code,
				Message: i18n.Translate(language, message),
				Data:    data,
			})
		} else {
			c.JSON(500, response)
		}
	}
	return
}

func Md5(str string) string {
	hash := md5.New()
	_, _ = io.WriteString(hash, str)
	return hex.EncodeToString(hash.Sum(nil))
}

func ValidateSignature(c *gin.Context) error {
	if !SignatureAuthEnabled() {
		return fmt.Errorf("api signature auth disabled")
	}

	query := c.Request.URL.Query()
	timestamp := query.Get("timestamp")
	sign := query.Get("sign")

	if timestamp == "" || sign == "" {
		return fmt.Errorf("missing api signature parameters")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid api timestamp")
	}
	now := timeNow().Unix()
	if absInt64(now-ts) > 60 {
		return fmt.Errorf("api signature expired")
	}

	body := ""
	if rawBody, exists := c.Get("rawBody"); exists {
		body, _ = rawBody.(string)
	}
	expectedSign := buildSignature(query, body)
	if !strings.EqualFold(expectedSign, sign) {
		return fmt.Errorf("invalid api signature")
	}
	return nil
}

func HasSignatureParams(c *gin.Context) bool {
	query := c.Request.URL.Query()
	return query.Get("timestamp") != "" || query.Get("sign") != ""
}

func SignatureAuthEnabled() bool {
	return strings.TrimSpace(config.Config.ApiSecret) != ""
}

func buildSignature(query url.Values, body string) string {
	queryParts := make([]string, 0, len(query))
	keys := make([]string, 0, len(query))
	for key := range query {
		if key == "sign" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		values := append([]string(nil), query[key]...)
		sort.Strings(values)
		for _, value := range values {
			queryParts = append(queryParts, key+"="+value)
		}
	}

	normalizedBody := normalizeBody(body)
	signSource := strings.Join(queryParts, "&") + "&body=" + normalizedBody + "&secret=" + config.Config.ApiSecret
	return Md5(signSource)
}

func normalizeBody(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, []byte(trimmed)); err == nil {
		return compacted.String()
	}
	return trimmed
}

func absInt64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
