package config

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	once        sync.Once
	configFiles string
)

func Load[T any](defaults map[string]any) *T {
	once.Do(func() {
		flag.StringVar(&configFiles, "config", "", "Файлы конфигурации через запятую (config.yaml,.env.local)")
		if !flag.Parsed() {
			flag.Parse()
		}
	})
	var res T
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	for file := range strings.SplitSeq(configFiles, ",") {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		if filepath.Ext(file) == ".env" || strings.HasPrefix(filepath.Base(file), ".env") {
			if err := godotenv.Load(file); err != nil {
				panic(fmt.Errorf("ошибка загрузки .env файла %s: %w", file, err))
			}
			continue
		}
		if filepath.Ext(file) == "" {
			v.SetConfigName(filepath.Base(file))
			v.AddConfigPath(filepath.Dir(file))
		} else {
			v.SetConfigFile(file)
		}
		if err := v.MergeInConfig(); err != nil {
			panic(fmt.Errorf("ошибка чтения файла конфигурации %s: %w", file, err))
		}
	}
	v.AutomaticEnv()
	if err := v.Unmarshal(&res); err != nil {
		panic(err)
	}
	return &res
}
