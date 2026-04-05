package router

import (
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
	v1.POST("/refresh-token", AdminAPI(api.RefreshToken, true))
	v1.Use(api.AuthMiddleware)

	// 数据源
	v1.POST("/data-sources", AdminAPI(api.NewDataSource, true))
	v1.POST("/data-sources/test", AdminAPI(api.TestDataSource, true))
	v1.GET("/data-sources", AdminAPI(api.GetDataSourceList))
	v1.GET("/data-sources/types", AdminAPI(api.GetDataSourceTypeList))
	v1.DELETE("/data-sources/:id", AdminAPI(api.DeleteDataSource))

	// 变量
	v1.GET("/variables", AdminAPI(api.GetVariableList))
	v1.GET("/variables/types", AdminAPI(api.GetVariableTypeList))
	v1.POST("/variables", AdminAPI(api.NewVariable))
	v1.DELETE("/variables/:id", AdminAPI(api.DeleteVariable))
	v1.POST("/variables/:id/test", AdminAPI(api.TestVariable))

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

	// 组件配置
	v1.GET("/components", AdminAPI(api.GetTypeByComponent))

	// 任务执行记录
	v1.GET("/task-records", AdminAPI(api.GetTaskRecordList))
	v1.POST("/task-records/:id/cancel", AdminAPI(api.CancelTaskRecord))
	v1.GET("/task-records/:id/files", AdminAPI(api.GetFileListByTaskRecordID))
	v1.GET("/task-records/:id/params", AdminAPI(api.GetTaskRecordParams))
	v1.GET("/task-records/:id/logs", AdminAPI(api.GetTaskRecordLogs))

	// 文件
	v1.GET("/files", AdminAPI(api.GetFileList))
	v1.POST("/files", AdminAPI(api.UploadFile, true))
	v1.DELETE("/files/:id", AdminAPI(api.DeleteFile))
}

func AdminAPI[T any](f func(*T, string) (interface{}, error), maskData ...bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetString("language")
		if len(maskData) > 0 && maskData[0] {
			c.Set("maskData", "true")
		}
		var req T
		if err := c.ShouldBind(&req); err != nil {
			c.Set("code", 1)
			c.Set("message", "invalid request parameters")
			c.Set("data", buildValidationErrorData(err, req, lang))
			c.Abort()
			return
		}
		// 绑定 URI 路径参数（如 :id），覆盖到同一个 req 结构体
		if err := c.ShouldBindUri(&req); err != nil {
			c.Set("code", 1)
			c.Set("message", "invalid request parameters")
			c.Set("data", buildValidationErrorData(err, req, lang))
			c.Abort()
			return
		}
		resp, err := f(&req, lang)
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
		} else {
			c.Set("code", 0)
			c.Set("data", resp)
			c.Set("message", "ok")
			return
		}
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
			uriTag := sf.Tag.Get("uri")
			if jsonTag != "" && jsonTag != "-" {
				fieldName = strings.Split(jsonTag, ",")[0]
			} else if uriTag != "" && uriTag != "-" {
				fieldName = strings.Split(uriTag, ",")[0]
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
