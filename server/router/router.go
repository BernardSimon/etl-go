package router

import (
	"net/http"

	"github.com/BernardSimon/etl-go/server/api"

	"github.com/gin-gonic/gin"
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

	// ── 旧路由（deprecated，保持向后兼容）────────────────────────────────────
	legacy := engine.Group("/etlApi")
	withMiddleware(legacy)
	legacy.POST("/login", func(c *gin.Context) {
		AdminAPI(api.LoginWithRateLimit(c), true)(c)
	})
	legacy.Use(api.AuthMiddleware)
	legacy.POST("/newDataSource", AdminAPI(api.NewDataSource, true))
	legacy.POST("/getDataSourceTypeList", AdminAPI(api.GetDataSourceTypeList))
	legacy.POST("/getDataSourceList", AdminAPI(api.GetDataSourceList))
	legacy.POST("/deleteDataSource", AdminAPI(api.DeleteDataSource))
	legacy.POST("/getVariableList", AdminAPI(api.GetVariableList))
	legacy.POST("/getVariableTypeList", AdminAPI(api.GetVariableTypeList))
	legacy.POST("/newVariable", AdminAPI(api.NewVariable))
	legacy.POST("/deleteVariable", AdminAPI(api.DeleteVariable))
	legacy.POST("/testVariable", AdminAPI(api.TestVariable))
	legacy.POST("/getTaskAll", AdminAPI(api.GetTaskAll))
	legacy.POST("/addTask", AdminAPI(api.AddTask))
	legacy.POST("/getTaskById", AdminAPI(api.GetTaskById))
	legacy.POST("/updateTask", AdminAPI(api.UpdateTask))
	legacy.POST("/runTask", AdminAPI(api.RunTask))
	legacy.POST("/deleteTask", AdminAPI(api.DeleteTask))
	legacy.POST("/stopTask", AdminAPI(api.StopTask))
	legacy.POST("/runTaskOnce", AdminAPI(api.RunTaskOnce))
	legacy.POST("/getTypeByComponent", AdminAPI(api.GetTypeByComponent))
	legacy.POST("/getTaskRecordList", AdminAPI(api.GetTaskRecordList))
	legacy.POST("/cancelTaskRecord", AdminAPI(api.CancelTaskRecord))
	legacy.POST("/getFileList", AdminAPI(api.GetFileList))
	legacy.POST("/uploadFile", AdminAPI(api.UploadFile, true))
	legacy.POST("/deleteFile", AdminAPI(api.DeleteFile))
	legacy.POST("/getFileListByTaskRecordID", AdminAPI(api.GetFileListByTaskRecordID))

	// ── v1 RESTful 路由 ───────────────────────────────────────────────────────
	v1 := engine.Group("/api/v1")
	withMiddleware(v1)
	v1.POST("/login", func(c *gin.Context) {
		AdminAPI(api.LoginWithRateLimit(c), true)(c)
	})
	v1.Use(api.AuthMiddleware)

	// 数据源
	v1.POST("/data-sources", AdminAPI(api.NewDataSource, true))
	v1.GET("/data-sources", AdminAPI(api.GetDataSourceList))
	v1.GET("/data-sources/types", AdminAPI(api.GetDataSourceTypeList))
	v1.DELETE("/data-sources", AdminAPI(api.DeleteDataSource))

	// 变量
	v1.GET("/variables", AdminAPI(api.GetVariableList))
	v1.GET("/variables/types", AdminAPI(api.GetVariableTypeList))
	v1.POST("/variables", AdminAPI(api.NewVariable))
	v1.DELETE("/variables", AdminAPI(api.DeleteVariable))
	v1.POST("/variables/test", AdminAPI(api.TestVariable))

	// 任务
	v1.GET("/tasks", AdminAPI(api.GetTaskAll))
	v1.POST("/tasks", AdminAPI(api.AddTask))
	v1.GET("/tasks/:id", AdminAPI(api.GetTaskById))
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

	// 文件
	v1.GET("/files", AdminAPI(api.GetFileList))
	v1.POST("/files", AdminAPI(api.UploadFile, true))
	v1.DELETE("/files", AdminAPI(api.DeleteFile))
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
			c.Set("message", "参数错误")
			c.Abort()
			return
		}
		resp, err := f(&req, lang)
		if err != nil {
			c.Set("code", 2)
			c.Set("message", err.Error())
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
