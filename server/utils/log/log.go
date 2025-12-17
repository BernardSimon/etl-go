package zapLog

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLog(isProduction bool) {
	var logger *zap.Logger

	// 创建 lumberjack 实例用于日志切割
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "./log/app.log", // 日志文件路径
		MaxSize:    20,              // 每个日志文件最大尺寸（MB）
		MaxBackups: 3,               // 最多保留3个备份
		MaxAge:     1,               // 最大保留天数
		Compress:   true,            // 是否压缩旧的日志文件
	}

	// 自定义时间格式和输出编码
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05")) // 设置时间格式
	}

	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// 设置日志级别
	var level zapcore.Level
	if isProduction {
		level = zapcore.InfoLevel // 生产环境只记录 Info 及以上级别
	} else {
		level = zapcore.DebugLevel // 测试环境记录 Debug 及以上级别
	}

	var core zapcore.Core
	if isProduction {
		// 生产环境只输出到文件，不输出到控制台
		fileWriter := zapcore.AddSync(lumberjackLogger)
		core = zapcore.NewCore(consoleEncoder, fileWriter, level)
	} else {
		// 测试环境同时输出到控制台和文件
		consoleDebugging := zapcore.Lock(os.Stdout)
		fileWriter := zapcore.AddSync(lumberjackLogger)
		core = zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, consoleDebugging, zapcore.WarnLevel),
			zapcore.NewCore(consoleEncoder, fileWriter, level),
		)
	}

	// 根据环境设置不同的配置
	if isProduction {
		logger = zap.New(core, zap.AddCaller())
	} else {
		logger = zap.New(core, zap.AddCaller(), zap.Development())
	}
	zap.ReplaceGlobals(logger)
}
