package logger

import (
	"go.uber.org/zap"
	"subservice/internal/config"
)

func New(cfgService *config.Config) (*zap.Logger, func()) {

	var cfg zap.Config
	if cfgService.Env == "prod" || cfgService.Env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	if lvl := cfgService.LogLevel; lvl != "" {
		_ = cfg.Level.UnmarshalText([]byte(lvl))
	}

	l, _ := cfg.Build()
	cleanup := func() {
		_ = l.Sync()
	}
	return l, cleanup
}
