package zapLog

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogOptions configures the logger. Zero values fall back to defaults.
type LogOptions struct {
	Filename   string // default: ./log/app.log
	MaxSize    int    // MB, default: 20
	MaxBackups int    // default: 3
	MaxAge     int    // days, default: 1
	Compress   bool
}

func defaultOptions(opt LogOptions) LogOptions {
	if opt.Filename == "" {
		opt.Filename = "./log/app.log"
	}
	if opt.MaxSize <= 0 {
		opt.MaxSize = 20
	}
	if opt.MaxBackups <= 0 {
		opt.MaxBackups = 3
	}
	if opt.MaxAge <= 0 {
		opt.MaxAge = 1
	}
	return opt
}

// NewLogger builds a *zap.Logger with the given options.
// Callers can pass the result to zap.ReplaceGlobals if desired.
func NewLogger(isProduction bool, opt LogOptions) *zap.Logger {
	opt = defaultOptions(opt)

	lumberjackLogger := &lumberjack.Logger{
		Filename:   opt.Filename,
		MaxSize:    opt.MaxSize,
		MaxBackups: opt.MaxBackups,
		MaxAge:     opt.MaxAge,
		Compress:   opt.Compress,
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	var level zapcore.Level
	if isProduction {
		level = zapcore.InfoLevel
	} else {
		level = zapcore.DebugLevel
	}

	fileWriter := zapcore.AddSync(lumberjackLogger)
	var core zapcore.Core
	if isProduction {
		core = zapcore.NewCore(encoder, fileWriter, level)
	} else {
		consoleWriter := zapcore.Lock(os.Stdout)
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, consoleWriter, zapcore.WarnLevel),
			zapcore.NewCore(encoder, fileWriter, level),
		)
	}

	if isProduction {
		return zap.New(core, zap.AddCaller())
	}
	return zap.New(core, zap.AddCaller(), zap.Development())
}

// InitLog is a convenience wrapper that builds a logger and sets it as the global logger.
func InitLog(isProduction bool, opt LogOptions) {
	logger := NewLogger(isProduction, opt)
	zap.ReplaceGlobals(logger)
}
