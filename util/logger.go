package util

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var SLogger *zap.SugaredLogger

func InitializeLogger(logLevel zapcore.Level) {
	config := zap.NewProductionConfig()
	config.Level.SetLevel(logLevel)
	Logger, _ = config.Build()
	SLogger = Logger.Sugar()
}
