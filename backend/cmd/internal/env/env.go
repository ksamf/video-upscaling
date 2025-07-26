package env

import (
	"os"
	"strconv"
	"strings"
)

type PgConfig struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}

type RedisConfig struct {
	Host string
	Port int
}

type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	EndpointURL     string
	BucketName      string
}

type Config struct {
	Postgres PgConfig
	Redis    RedisConfig
	S3       S3Config
}

func New() *Config {
	return &Config{
		Postgres: PgConfig{
			Host: getEnv("DB_HOST", ""),
			Port: getEnvAsInt("DB_PORT", 0),
			User: getEnv("DB_USER", ""),
			Pass: getEnv("DB_PASS", ""),
			Name: getEnv("DB_NAME", ""),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", ""),
			Port: getEnvAsInt("REDIS_PORT", 0),
		},
		S3: S3Config{
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			EndpointURL:     getEnv("ENDPOINT_URL", ""),
			BucketName:      getEnv("BUCKET_NAME", ""),
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
