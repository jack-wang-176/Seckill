package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局日志对象，后续可以用 DI (依赖注入) 替换掉它
var Log *zap.Logger

// InitLogger 初始化 Zap 日志
func InitLogger() *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // 可读的时间格式
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 控制台输出带颜色的级别

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // 终端输出格式
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)
	// AddCaller() 可以在日志中打印出是哪个文件哪一行打印的日志
	Log = zap.New(core, zap.AddCaller())
	return Log
}
