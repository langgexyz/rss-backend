package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Fetcher  FetcherConfig
}

type ServerConfig struct {
	Port        int
	CORSOrigins []string `mapstructure:"cors_origins"`
}

type DatabaseConfig struct {
	DSN string
}

type FetcherConfig struct {
	TimeoutSeconds     int `mapstructure:"timeout_seconds"`
	MaxArticlesPerFeed int `mapstructure:"max_articles_per_feed"`
	MaxConcurrent      int `mapstructure:"max_concurrent"`
	MinRefreshInterval int `mapstructure:"min_refresh_interval"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, viper.Unmarshal(&cfg)
}
