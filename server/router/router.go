package router

import (
	"net/http"

	"github.com/BernardSimon/etl-go/server/api"

	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine) {
	fileRouter := engine.Group("/file")
	fileRouter.Use(api.AuthMiddlewareFile) // 添加认证中间件
	fileRouter.StaticFS("/", http.Dir("./file"))
	admin := engine.Group("/etlApi")
	admin.Use(api.RequestResponseMiddleware)
	admin.POST("/login", AdminAPI(api.Login, true))
	admin.Use(api.AuthMiddleware)
	admin.POST("/newDataSource", AdminAPI(api.NewDataSource, true))
	admin.POST("/getDataSourceTypeList", AdminAPI(api.GetDataSourceTypeList))
	admin.POST("/getDataSourceList", AdminAPI(api.GetDataSourceList))
	admin.POST("/deleteDataSource", AdminAPI(api.DeleteDataSource))
	admin.POST("/getVariableList", AdminAPI(api.GetVariableList))
	admin.POST("/getVariableTypeList", AdminAPI(api.GetVariableTypeList))
	admin.POST("/newVariable", AdminAPI(api.NewVariable))
	admin.POST("/deleteVariable", AdminAPI(api.DeleteVariable))
	admin.POST("/testVariable", AdminAPI(api.TestVariable))
	admin.POST("/getTaskAll", AdminAPI(api.GetTaskAll))
	admin.POST("/addTask", AdminAPI(api.AddTask))
	admin.POST("/getTaskById", AdminAPI(api.GetTaskById))
	admin.POST("/updateTask", AdminAPI(api.UpdateTask))
	admin.POST("/runTask", AdminAPI(api.RunTask))
	admin.POST("/deleteTask", AdminAPI(api.DeleteTask))
	admin.POST("/stopTask", AdminAPI(api.StopTask))
	admin.POST("/runTaskOnce", AdminAPI(api.RunTaskOnce))
	admin.POST("/getTypeByComponent", AdminAPI(api.GetTypeByComponent))
	admin.POST("/getTaskRecordList", AdminAPI(api.GetTaskRecordList))
	admin.POST("/cancelTaskRecord", AdminAPI(api.CancelTaskRecord))
	admin.POST("/getFileList", AdminAPI(api.GetFileList))
	admin.POST("/uploadFile", AdminAPI(api.UploadFile, true))
	admin.POST("/deleteFile", AdminAPI(api.DeleteFile))
	admin.POST("/getFileListByTaskRecordID", AdminAPI(api.GetFileListByTaskRecordID))
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
