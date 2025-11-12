package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Images       ImagesConfig       `mapstructure:"images"`
	Experimental ExperimentalConfig `mapstructure:"experimental"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port              int               `mapstructure:"port"`
	EnablePrintRoutes bool              `mapstructure:"enable_print_routes"`
	ReadTimeout       time.Duration     `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration     `mapstructure:"write_timeout"`
	Limiter           LimiterConfig     `mapstructure:"limiter"`
	CrossOriginPolicy CrossOriginPolicy `mapstructure:"cross_origin_policy"`
}

// CrossOriginPolicy holds cross-origin policy configuration
type CrossOriginPolicy struct {
	Corp             string `mapstructure:"corp"`
	CorsAllowOrigins string `mapstructure:"cors_allow_origins"`
	CorsAllowMethods string `mapstructure:"cors_allow_methods"`
	CorsAllowHeaders string `mapstructure:"cors_allow_headers"`
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

// ExperimentalConfig holds experimental feature configuration
type ExperimentalConfig struct {
	EnableUploadAPI bool `mapstructure:"enable_upload_api"`
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

	viper.SetDefault("server.cross_origin_policy.corp", "cross-origin")
	viper.SetDefault("server.cross_origin_policy.cors_allow_origins", "*")
	viper.SetDefault("server.cross_origin_policy.cors_allow_methods", "GET,POST,PUT,DELETE,HEAD,OPTIONS")
	viper.SetDefault("server.cross_origin_policy.cors_allow_headers", "Origin,Content-Type,Accept,Authorization,Cache-Control,If-None-Match")

	// Images defaults
	viper.SetDefault("images.base_path", "web/static/assets")
	viper.SetDefault("max_dimension", 8192)                  // 8K
	viper.SetDefault("max_file_size_in_bytes", 50*1024*1024) // 50MB

	// Rate limiter defaults
	viper.SetDefault("server.limiter.max", 100)
	viper.SetDefault("server.limiter.expiration", "1m")

	// Experimental defaults
	viper.SetDefault("experimental.enable_upload_api", false)
}
