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

	AuthConfig struct {
		JWTSecret       string `toml:"jwt_secret"`
		JWTAccessExpiry int    `toml:"jwt_access_expiry"`
	} `toml:"auth"`
}

type Config struct {
	APIKeys map[string]string

	Database DBConfig
	Server   struct {
		Host    string
		Port    int
		IsDevel bool
	}
	AuthConfig struct {
		JWTSecret       string
		JWTAccessExpiry int // in hours
	}
}

var (
	appConfig *AppConfig
	loadOnce  sync.Once
	errLoad   error
)

func Load(path string) (Config, error) {
	loadOnce.Do(func() {
		var cfg AppConfig
		if _, errLoad = toml.DecodeFile(path, &cfg); errLoad != nil {
			errLoad = fmt.Errorf("failed to load config: %w", errLoad)
			return
		}
		appConfig = &cfg
	})

	if errLoad != nil {
		return Config{}, errLoad
	}

	cfg := Config{
		APIKeys:  appConfig.APIKeys,
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
		AuthConfig: struct {
			JWTSecret       string
			JWTAccessExpiry int
		}{
			JWTSecret:       appConfig.AuthConfig.JWTSecret,
			JWTAccessExpiry: appConfig.AuthConfig.JWTAccessExpiry,
		},
	}

	return cfg, nil
}

func GetAppConfig() *AppConfig {
	return appConfig
}
