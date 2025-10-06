package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type PgConfig struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}
type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	EndpointURL     string
	BucketName      string
}
type Config struct {
	Postgres PgConfig
	S3       S3Config
}

func New() *Config {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
	return &Config{
		Postgres: PgConfig{
			Host: getEnv("DB_HOST", ""),
			Port: getEnvAsInt("DB_PORT", 5432),
			User: getEnv("DB_USER", ""),
			Pass: getEnv("DB_PASS", ""),
			Name: getEnv("DB_NAME", ""),
		},
		S3: S3Config{
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
			EndpointURL:     getEnv("S3_ENDPOINT_URL", ""),
			BucketName:      getEnv("S3_BUCKET_NAME", ""),
		},
	}
}
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
