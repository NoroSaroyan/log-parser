package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger(level string) error {
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(zap.InfoLevel)

	l, err := cfg.Build()
	if err != nil {
		return err
	}
	Logger = l
	return nil
}
