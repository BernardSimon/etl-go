package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/BernardSimon/etl-go/server/config"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/BernardSimon/etl-go/server/router"
	"github.com/BernardSimon/etl-go/server/task"
	zapLog "github.com/BernardSimon/etl-go/server/utils/log"

	"gorm.io/gorm/schema"

	_ "github.com/BernardSimon/etl-go/etl" //注册 etl 组件
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:embed web/dist/*
//go:embed web/dist/assets/*
var staticFiles embed.FS

func main() {
	created, err := config.EnsureConfig()
	if err != nil {
		fmt.Printf("Failed to initialize config: %v\n", err)
		os.Exit(1)
	}
	if created {
		fmt.Println("Config file not found. A new config.yaml has been initialized with secure defaults.")
	}
	if err := config.EnsureRuntimePaths(); err != nil {
		fmt.Printf("Failed to prepare runtime directories: %v\n", err)
		os.Exit(1)
	}
	// 弱密码警告
	if config.Config.Password == "password123" {
		fmt.Println("WARNING: Using default weak password. Set ETL_PASSWORD environment variable or update config.yaml.")
	}
	// 日志初始化
	logOpt := zapLog.LogOptions{
		Filename:   config.Config.Log.Filename,
		MaxSize:    config.Config.Log.MaxSize,
		MaxBackups: config.Config.Log.MaxBackups,
		MaxAge:     config.Config.Log.MaxAge,
		Compress:   config.Config.Log.Compress,
	}
	switch config.Config.LogLevel {
	case "prod":
		gin.SetMode(gin.ReleaseMode)
		zapLog.InitLog(true, logOpt)
	default:
		gin.SetMode(gin.DebugMode)
		zapLog.InitLog(false, logOpt)
	}
	// 数据库自定义序列化-加密字段
	schema.RegisterSerializer("encryption", &model.EncryptionSerializer{})
	// 数据库连接
	if err := model.InitDb(); err != nil {
		zap.L().Fatal("Failed To Connect Database", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(err))
	}
	// 是否迁移数据库
	if config.Config.InitDb {
		err := model.MigrateDb()
		if err != nil {
			zap.L().Fatal("Failed To Migrate Database", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(err))
		}
		config.Config.InitDb = false
		err = config.SaveConfig()
		if err != nil {
			zap.L().Fatal("Failed To Save Config", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(err))
		}
		zap.L().Info("Database Migration Success", zap.String("service", "system"), zap.String("name", config.Ip))
	}
	startService(config.Config.RunWeb)
	select {}
}

func startService(startWeb bool) {
	var servers []*http.Server
	if startWeb {
		q := gin.New()
		// 提供静态文件服务 - 使用嵌入的文件系统
		embeddedDist, err := fs.Sub(staticFiles, "web/dist/assets")
		if err != nil {
			zap.L().Fatal("Failed To Create Sub Filesystem", zap.Error(err))
		}
		q.StaticFS("/assets", http.FS(embeddedDist))
		// 更健壮的 NoRoute 处理
		q.NoRoute(func(c *gin.Context) {
			// 如果是前端路由，返回 index.html
			indexFile, err := staticFiles.ReadFile("web/dist/index.html")
			if err != nil {
				zap.L().Error("Failed To Read index.html", zap.Error(err), zap.String("service", "system"), zap.String("name", config.Ip))
				c.AbortWithStatus(404)
				return
			}
			c.Data(200, "text/html; charset=utf-8", indexFile)
		})
		// 模拟nginx启动前端服务
		srvWeb := &http.Server{
			Addr:         config.Config.WebUrl,
			Handler:      q,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		servers = append(servers, srvWeb)
		// 启动前端服务
		go func() {
			err := srvWeb.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				zap.L().Fatal(fmt.Sprintf("Failed To Start Web Server"), zap.Error(err), zap.String("service", "system"), zap.String("name", config.Ip))
			}
		}()
		zap.L().Info(fmt.Sprintf("Web Server Started At %s", config.Config.WebUrl), zap.String("service", "system"), zap.String("name", config.Ip))
		// 尝试打开前端网页
		openBrowser("http://" + config.Config.WebUrl)
	}
	r := gin.New()
	// Keep multipart form memory bounded; larger uploads spill to temp files.
	r.MaxMultipartMemory = 8 << 20
	// 全局中间件 - CORS 配置
	corsOrigins := config.Config.CorsOrigins
	if len(corsOrigins) == 0 {
		// 未配置时默认允许 web 服务地址
		corsOrigins = []string{"http://" + config.Config.WebUrl}
	}
	allowCredentials := true
	// 如果显式配置了 "*"，则不能同时启用 credentials
	for _, origin := range corsOrigins {
		if origin == "*" {
			allowCredentials = false
			break
		}
	}
	var corsConfig = cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept-Language"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowCredentials,
		MaxAge:           12 * time.Hour,
	}
	var Cors = cors.New(corsConfig)
	r.Use(Cors)

	// 路由注册
	router.Register(r)
	zap.L().Info("Router Registration Complete", zap.String("service", "system"), zap.String("name", config.Ip))

	// 配置HTTP服务器
	srv := &http.Server{
		Addr:         config.Config.ServerUrl,
		Handler:      r,
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}
	servers = append(servers, srv)
	// 启动后端服务
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal(fmt.Sprintf("Failed To Start Api Server"), zap.Error(err), zap.String("service", "system"), zap.String("name", config.Ip))
		}
	}()
	// 优雅关机处理
	go gracefulShutdown(srv)
	zap.L().Info(fmt.Sprintf("Api Server Started At %s", config.Config.ServerUrl), zap.String("service", "system"), zap.String("name", config.Ip))
	// 启动计划任务
	task.SetMissions()
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		winUrl := strings.Replace(url, "0.0.0.0", "127.0.0.1", -1)
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", winUrl).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		zap.L().Warn(fmt.Sprintf("Unknown OS: %s, Failed To Init Web Url", runtime.GOOS), zap.String("service", "system"), zap.String("name", config.Ip))
		return
	}
	if err != nil {
		zap.L().Warn("Failed To Init Web Url", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(err))
	} else {
		zap.L().Info("Init Web Url Success", zap.String("service", "system"), zap.String("name", config.Ip))
	}
}

func gracefulShutdown(servers ...*http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Server is shutting down", zap.String("service", "system"), zap.String("name", config.Ip))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutDown := false
	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			shutDown = true
			zap.L().Error(fmt.Sprintf("Failed to stop server %s", srv.Addr), zap.Error(err), zap.String("service", "system"), zap.String("name", config.Ip))
		}
	}
	if shutDown {
		zap.L().Fatal("Servers shutdown forcefully", zap.String("service", "system"), zap.String("name", config.Ip))
	}
	zap.L().Info("Servers terminated successfully", zap.String("service", "system"), zap.String("name", config.Ip))
	// Sync logs
	_ = zap.L().Sync()
	os.Exit(0)
}
