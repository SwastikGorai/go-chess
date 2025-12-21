package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
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
	SSLMode  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

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
			SSLMode:  GetEnv("DB_SSLMODE", "disable").(string),
		},
	}

	return cfg, nil
}

func (db DatabaseConfig) DSN() string {
	userInfo := url.User(db.User)
	if db.Password != "" {
		userInfo = url.UserPassword(db.User, db.Password)
	}

	u := url.URL{
		Scheme: "postgres",
		User:   userInfo,
		Host:   fmt.Sprintf("%s:%d", db.Host, db.Port),
		Path:   db.DBName,
	}

	q := u.Query()
	if db.SSLMode != "" {
		q.Set("sslmode", db.SSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
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
