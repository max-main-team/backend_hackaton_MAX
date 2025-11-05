package config

import (
	"fmt"
	"sync"

	"github.com/BurntSushi/toml"
)

type DBConfig struct {
	ConnString string `toml:"conn_string"`
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	Database   string `toml:"database"`
	SSLMode    string `toml:"sslmode"`
}

type AppConfig struct {
	APIKeys  map[string]string `toml:"api_keys"`
	Database DBConfig          `toml:"database"`
	Server   struct {
		Host    string `toml:"host"`
		Port    int    `toml:"port"`
		IsDevel bool   `toml:"is_devel"`
	} `toml:"server"`
}

type Config struct {
	Database DBConfig
	Server   struct {
		Host    string
		Port    int
		IsDevel bool
	}
}

var (
	appConfig *AppConfig
	loadOnce  sync.Once
	loadErr   error
)

func Load(path string) (Config, error) {
	loadOnce.Do(func() {
		var cfg AppConfig
		if _, loadErr = toml.DecodeFile(path, &cfg); loadErr != nil {
			loadErr = fmt.Errorf("failed to load config: %w", loadErr)
			return
		}
		appConfig = &cfg
	})

	if loadErr != nil {
		return Config{}, loadErr
	}

	cfg := Config{
		Database: appConfig.Database,
		Server: struct {
			Host    string
			Port    int
			IsDevel bool
		}{
			Host:    appConfig.Server.Host,
			Port:    appConfig.Server.Port,
			IsDevel: appConfig.Server.IsDevel,
		},
	}

	return cfg, nil
}

func GetAppConfig() *AppConfig {
	return appConfig
}
