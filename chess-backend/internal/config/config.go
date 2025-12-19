package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: GetEnv("PORT", 8080).(int),
		},
		Database: DatabaseConfig{
			Host:     GetEnv("DB_HOST", "localhost").(string),
			Port:     GetEnv("DB_PORT", 5432).(int),
			User:     GetEnv("DB_USER", "chess").(string),
			Password: GetEnv("DB_PASSWORD", "").(string),
			DBName:   GetEnv("DB_NAME", "chess_db").(string),
		},
	}

	return cfg, nil
}

func GetEnv(key string, defaultValue any) any {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	switch def := defaultValue.(type) {
	case string:
		return value
	case int:
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		return def

	case bool:
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		return def
	default:
		panic(fmt.Sprintf("unsupported type %T", defaultValue))

	}
}
