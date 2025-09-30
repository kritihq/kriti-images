package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Images  ImagesConfig  `mapstructure:"images"`
	Limiter LimiterConfig `mapstructure:"limiter"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port              int           `mapstructure:"port"`
	EnablePrintRoutes bool          `mapstructure:"enable_print_routes"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
}

// ImagesConfig holds image-specific configuration
type ImagesConfig struct {
	BasePath            string `mapstructure:"base_path"`
	MaxImageDimension   int    `mapstructure:"max_image_dimension"`
	MaxImageSizeInBytes int64  `mapstructure:"max_file_size_in_bytes"`
}

// LimiterConfig holds rate limiter configuration
type LimiterConfig struct {
	Max        int           `mapstructure:"max"`
	Expiration time.Duration `mapstructure:"expiration"`
}

// LoadConfig reads configuration from file and environment variables
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Read config file (supports yaml, toml, json, etc.)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.enable_print_routes", false)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	// Images defaults
	viper.SetDefault("images.base_path", "web/static/assets")
	viper.SetDefault("max_dimension", 8192)                  // 8K
	viper.SetDefault("max_file_size_in_bytes", 50*1024*1024) // 50MB

	// Rate limiter defaults
	viper.SetDefault("limiter.max", 100)
	viper.SetDefault("limiter.expiration", "1m")
}
