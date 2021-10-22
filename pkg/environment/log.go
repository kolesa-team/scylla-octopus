package environment

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogOptions struct {
	Level zapcore.Level
}

func getLogger(options LogOptions) *zap.SugaredLogger {
	logCfg := zap.NewDevelopmentConfig()
	logCfg.EncoderConfig.TimeKey = zapcore.OmitKey
	logCfg.EncoderConfig.CallerKey = zapcore.OmitKey
	logCfg.EncoderConfig.ConsoleSeparator = " "
	logCfg.Level = zap.NewAtomicLevelAt(options.Level)

	logger, _ := logCfg.Build()

	return logger.Sugar()
}
