package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv           string
	Port             string
	DatabaseURL      string
	StorageRoot      string
	PublicBaseURL    string
	AllowedOrigins   []string
	LogLevel         string
	XelatexPath      string
	MagickPath       string
	// PDF conversion settings
	PdfConverterType string // "local", "mineru-cloud", "disabled"
	MinerUAPIKey     string
	MinerUAPIBase    string
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigType("env")
	v.AutomaticEnv()

	v.SetDefault("APP_ENV", "development")
	v.SetDefault("PORT", "8080")
	v.SetDefault("DATABASE_URL", "postgres://mathlib:mathlib@localhost:5432/mathlib?sslmode=disable")
	v.SetDefault("STORAGE_ROOT", "./storage")
	v.SetDefault("PUBLIC_BASE_URL", "http://localhost:8080")
	v.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000")
	v.SetDefault("LOG_LEVEL", "debug")
	v.SetDefault("XELATEX_PATH", "xelatex")
	v.SetDefault("MAGICK_PATH", "convert")
	// PDF conversion defaults
	v.SetDefault("PDF_CONVERTER_TYPE", "local")
	v.SetDefault("MINERU_API_BASE", "https://mineru.net/api/v4")

	if envPath := os.Getenv("SERVER_ENV_PATH"); envPath != "" {
		v.SetConfigFile(envPath)
	} else {
		v.SetConfigName(".env")
		v.AddConfigPath(".")
		v.AddConfigPath("./server")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := Config{
		AppEnv:           v.GetString("APP_ENV"),
		Port:             v.GetString("PORT"),
		DatabaseURL:      v.GetString("DATABASE_URL"),
		StorageRoot:      v.GetString("STORAGE_ROOT"),
		PublicBaseURL:    strings.TrimRight(v.GetString("PUBLIC_BASE_URL"), "/"),
		AllowedOrigins:   splitCSV(v.GetString("ALLOWED_ORIGINS")),
		LogLevel:         v.GetString("LOG_LEVEL"),
		XelatexPath:      v.GetString("XELATEX_PATH"),
		MagickPath:       v.GetString("MAGICK_PATH"),
		PdfConverterType: v.GetString("PDF_CONVERTER_TYPE"),
		MinerUAPIKey:     v.GetString("MINERU_API_KEY"),
		MinerUAPIBase:    v.GetString("MINERU_API_BASE"),
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	return nil
}

func (c Config) AdminAPIKey() string {
	return c.getEnv("ADMIN_API_KEY", "")
}

func (c Config) getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
