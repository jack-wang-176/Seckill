package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var apiLogger *zap.Logger

func InitLogger() *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)

	lg := zap.New(core, zap.AddCaller())
	apiLogger = lg
	return lg
}

func SetLogger(l *zap.Logger) {
	apiLogger = l
}

func GetLogger() *zap.Logger {
	return apiLogger
}
