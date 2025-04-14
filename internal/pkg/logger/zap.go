package logger

import (
	"github.com/maksemen2/pvz-service/config"
	"go.uber.org/zap"
)

// NewZapLogger создает новый экземпляр логгера с использованием конфигурации.
// Поле cfg.Level должно быть одним из: "debug", "info", "warn", "error" или "silent".
// При уровне "silent" возвращается пустой логгер (zap.NewNop).
func NewZapLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()

	switch cfg.Level {
	case "debug":
		zapConfig.Level.SetLevel(zap.DebugLevel)
	case "info":
		zapConfig.Level.SetLevel(zap.InfoLevel)
	case "warn":
		zapConfig.Level.SetLevel(zap.WarnLevel)
	case "error":
		zapConfig.Level.SetLevel(zap.ErrorLevel)
	case "silent":
		return zap.NewNop(), nil
	}

	return zapConfig.Build()
}
