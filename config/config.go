// Package config provides application configuration via YAML or environment variables.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Meilisearch MeilisearchConfig `mapstructure:"meilisearch"`
	PDPL        PDPLConfig        `mapstructure:"pdpl"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type MeilisearchConfig struct {
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"api_key"`
	Index  string `mapstructure:"index"`
}

type PDPLConfig struct {
	RetentionDays int `mapstructure:"retention_days"`
	AnalyticsDays int `mapstructure:"analytics_days"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	v.SetConfigFile(path)

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("meilisearch.host", "http://localhost:7700")
	v.SetDefault("meilisearch.index", "uae-search")
	v.SetDefault("pdpl.retention_days", 90)
	v.SetDefault("pdpl.analytics_days", 365)

	v.SetEnvPrefix("SEARCHJET")
	v.AutomaticEnv()

	if path != "" && path != "SEARCHJET_ENV" {
		if _, err := os.Stat(path); err == nil {
			if err := v.ReadInConfig(); err != nil {
				return nil, err
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func DefaultRetention() time.Duration {
	return time.Duration(90) * 24 * time.Hour
}

func DefaultAnalyticsRetention() time.Duration {
	return time.Duration(365) * 24 * time.Hour
}
