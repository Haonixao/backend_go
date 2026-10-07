package config

import (
	"sync"

	"backend_go/pkg/config"
	"backend_go/pkg/gorm_extra"
	"backend_go/pkg/log"
	"backend_go/pkg/utils/maps"
)

type Config struct {
	*log.Log `mapstructure:"log"`

	*gorm_extra.Postgres `mapstructure:"postgres"`

	App struct {
		Mode             string `mapstructure:"mode"`
		Port             string `mapstructure:"port"`
		SchedulerEnabled bool   `mapstructure:"scheduler_enabled"`
		KeyForCryptUtils string `mapstructure:"key_for_crypt_utils"`
		ServiceUrl       string `mapstructure:"service_url"`
	} `mapstructure:"app"`

	Cors struct {
		AllowedOrigins string `mapstructure:"allowed_origins"` // через запятую
	} `mapstructure:"cors"`
}

var defaults = map[string]any{
	"app.mode":                "debug",
	"app.port":                "8080",
	"app.scheduler_enabled":   true,
	"app.key_for_crypt_utils": "6368616e676520746869732070617373",
	"app.service_url":         "http://localhost",

	"cors.allowed_origins": "http://localhost",
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		defaults = maps.Merge(
			defaults,
			log.Defaults,
			gorm_extra.PostgresDefaults,
		)
		defaults["postgres.db"] = "main"
		instance = config.Load[Config](defaults)
		if instance.App.ServiceUrl == "http://localhost" {
			instance.App.ServiceUrl = "http://localhost:" + instance.App.Port
		}
	})
	return instance
}
