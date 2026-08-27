package router

import (
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/BernardSimon/etl-go/server/api"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Register(engine *gin.Engine) {
	// 文件静态服务（带 token 鉴权）
	fileRouter := engine.Group("/file")
	fileRouter.Use(api.AuthMiddlewareFile)
	fileRouter.StaticFS("/", http.Dir("./file"))

	// 公共中间件（日志、token 提取、语言）
	withMiddleware := func(g *gin.RouterGroup) {
		g.Use(api.RequestResponseMiddleware)
	}

	// ── v1 RESTful 路由 ───────────────────────────────────────────────────────
	v1 := engine.Group("/api/v1")
	withMiddleware(v1)
	v1.POST("/login", func(c *gin.Context) {
		AdminAPI(api.LoginWithRateLimit(c), true)(c)
	})
	v1.POST("/verify-2fa", func(c *gin.Context) {
		AdminAPI(api.VerifyTwoFactorWithRateLimit(c), true)(c)
	})
	v1.POST("/refresh-token", AdminAPI(api.RefreshToken))
	v1.Use(api.AuthMiddleware)

	// 数据源
	v1.POST("/data-sources", AdminAPI(api.NewDataSource, true))
	v1.POST("/data-sources/test", AdminAPI(api.TestDataSource, true))
	v1.GET("/data-sources", AdminAPI(api.GetDataSourceList))
	v1.GET("/data-sources/types", AdminAPI(api.GetDataSourceTypeList))
	v1.GET("/data-sources/:id", AdminAPI(api.GetDataSourceById))
	v1.GET("/data-sources/:id/schema", AdminAPI(api.GetDataSourceSchema))
	v1.DELETE("/data-sources/:id", AdminAPI(api.DeleteDataSource))

	// 变量
	v1.GET("/variables", AdminAPI(api.GetVariableList))
	v1.GET("/variables/types", AdminAPI(api.GetVariableTypeList))
	v1.POST("/variables", AdminAPI(api.NewVariable))
	v1.DELETE("/variables/:id", AdminAPI(api.DeleteVariable))
	v1.POST("/variables/:id/test", AdminAPI(api.TestVariable))

	// 标签
	v1.GET("/tags", AdminAPI(api.GetTagList))
	v1.POST("/tags", AdminAPI(api.AddTag))
	v1.PUT("/tags/:id", AdminAPI(api.UpdateTag))
	v1.DELETE("/tags/:id", AdminAPI(api.DeleteTag))

	// 任务
	v1.GET("/tasks", AdminAPI(api.GetTaskAll))
	v1.POST("/tasks", AdminAPI(api.AddTask))
	v1.GET("/task-templates", AdminAPI(api.GetTaskTemplateList))
	v1.POST("/task-templates", AdminAPI(api.SaveTaskTemplate))
	v1.DELETE("/task-templates/:id", AdminAPI(api.DeleteTaskTemplate))
	v1.GET("/tasks/:id", AdminAPI(api.GetTaskById))
	v1.GET("/tasks/:id/records", AdminAPI(api.GetTaskRecordsByTaskID))
	v1.GET("/tasks/:id/log", AdminAPI(api.GetTaskLatestLog))
	v1.PUT("/tasks/:id", AdminAPI(api.UpdateTask))
	v1.DELETE("/tasks/:id", AdminAPI(api.DeleteTask))
	v1.POST("/tasks/:id/schedule", AdminAPI(api.RunTask))
	v1.POST("/tasks/:id/stop", AdminAPI(api.StopTask))
	v1.POST("/tasks/:id/run", AdminAPI(api.RunTaskOnce))
	v1.POST("/tasks/:id/preview", AdminAPI(api.PreviewTask))

	// 组件配置
	v1.GET("/components", AdminAPI(api.GetTypeByComponent))

	// 任务执行记录
	v1.GET("/task-records", AdminAPI(api.GetTaskRecordList))
	v1.DELETE("/task-records", AdminAPI(api.CleanTaskRecords))
	v1.POST("/task-records/:id/cancel", AdminAPI(api.CancelTaskRecord))
	v1.GET("/task-records/:id/files", AdminAPI(api.GetFileListByTaskRecordID))
	v1.GET("/task-records/:id/params", AdminAPI(api.GetTaskRecordParams))
	v1.GET("/task-records/:id/logs", AdminAPI(api.GetTaskRecordLogs))

	// 文件
	v1.GET("/files", AdminAPI(api.GetFileList))
	v1.POST("/files", AdminAPI(api.UploadFile, true))
	v1.DELETE("/files/:id", AdminAPI(api.DeleteFile))

	// 分片上传会话管理（JSON body，走完整中间件）
	v1.POST("/files/upload/init", AdminAPI(api.InitUploadSession))
	v1.GET("/files/upload/:session_id", AdminAPI(api.GetUploadStatus))
	v1.POST("/files/upload/:session_id/complete", AdminAPI(api.CompleteUpload))
	v1.DELETE("/files/upload/:session_id", AdminAPI(api.CancelUpload))

	// 分片数据上传（二进制流，绕过 body 预读中间件，由 handler 自行鉴权）
	v1Raw := engine.Group("/api/v1")
	v1Raw.PUT("/files/upload/:session_id/chunk/:chunk_index", api.UploadChunkRaw)
}

// AdminAPI 是通用路由处理器工厂，URI 参数和 Body 参数完全分离：
//   - URI: c.ShouldBindUri 绑定（含 validator），URI 结构体只含 uri tag，与 body 无交叉
//   - Body: c.ShouldBind 绑定（GET→query form，POST/PUT→JSON body），含 validator
func AdminAPI[URI, Body any](f func(*URI, *Body, string) (interface{}, error), maskData ...bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetString("language")
		if len(maskData) > 0 && maskData[0] {
			c.Set("maskData", "true")
		}

		// 第一步：URI 参数绑定（含 validator）
		// URI 结构体只含 uri tag 字段，与 body 完全隔离，validator 不会误判
		var uri URI
		if err := c.ShouldBindUri(&uri); err != nil {
			c.Set("code", 1)
			c.Set("message", "invalid request parameters")
			c.Set("data", buildValidationErrorData(err, uri, lang))
			c.Abort()
			return
		}

		// 第二步：Body/Query 参数绑定 + validator
		var body Body
		if err := c.ShouldBind(&body); err != nil && err != io.EOF {
			c.Set("code", 1)
			c.Set("message", "invalid request parameters")
			c.Set("data", buildValidationErrorData(err, body, lang))
			c.Abort()
			return
		}

		resp, err := f(&uri, &body, lang)
		if err != nil {
			if serviceErr, ok := err.(*types.ServiceError); ok {
				c.Set("code", serviceErr.Code)
				c.Set("message", serviceErr.Message)
				if serviceErr.Data != nil {
					c.Set("data", serviceErr.Data)
				}
			} else {
				c.Set("code", 2)
				c.Set("message", err.Error())
			}
			c.Abort()
			return
		}
		c.Set("code", 0)
		c.Set("data", resp)
		c.Set("message", "ok")
	}
}

func buildValidationErrorData[T any](err error, req T, lang string) types.ErrorData {
	data := types.ErrorData{}
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return data
	}

	reqType := reflect.TypeOf(req)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	for _, fieldErr := range validationErrors {
		fieldName := fieldErr.Field()
		if sf, ok := reqType.FieldByName(fieldName); ok {
			jsonTag := sf.Tag.Get("json")
			formTag := sf.Tag.Get("form")
			if jsonTag != "" && jsonTag != "-" {
				fieldName = strings.Split(jsonTag, ",")[0]
			} else if formTag != "" && formTag != "-" {
				fieldName = strings.Split(formTag, ",")[0]
			}
		}
		message := "invalid field"
		if fieldErr.Tag() == "required" {
			message = i18n.Translate(lang, "field is required")
		} else {
			message = i18n.Translate(lang, "invalid field")
		}
		data.Errors = append(data.Errors, types.FieldError{
			Field:   fieldName,
			Message: message,
		})
	}
	return data
}
