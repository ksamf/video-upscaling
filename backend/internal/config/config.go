package config

import (
	"os"
	"strconv"
)

type AppConfig struct {
	Host  string
	Port  int
	Debug string
}
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
	Pass string
}
type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	EndpointURL     string
	BucketName      string
}

type ApiConfig struct {
	BaseURL string
}

type KafkaConfig struct {
	Host string
	Port int
}
type Config struct {
	App      AppConfig
	Postgres PgConfig
	Redis    RedisConfig
	S3       S3Config
	Api      ApiConfig
	Kafka    KafkaConfig
}

func New() *Config {
	// if err := godotenv.Load(".env.dev"); err != nil {
	// 	panic(err)
	// }
	return &Config{
		App: AppConfig{
			Host:  getEnv("APP_HOST", "localhost"),
			Port:  getEnvAsInt("APP_PORT", 8000),
			Debug: getEnv("APP_DEBUG", "release"),
		},
		Postgres: PgConfig{
			Host: getEnv("DB_HOST", ""),
			Port: getEnvAsInt("DB_PORT", 5432),
			User: getEnv("DB_USER", ""),
			Pass: getEnv("DB_PASS", ""),
			Name: getEnv("DB_NAME", ""),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnvAsInt("REDIS_PORT", 6379),
			Pass: getEnv("REDIS_PASS", ""),
		},
		S3: S3Config{
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
			EndpointURL:     getEnv("S3_ENDPOINT_URL", ""),
			BucketName:      getEnv("S3_BUCKET_NAME", ""),
		},
		Api: ApiConfig{
			BaseURL: getEnv("BASE_URL", ""),
		},
		Kafka: KafkaConfig{
			Host: getEnv("KAFKA_HOST", "localhost"),
			Port: getEnvAsInt("KAFKA_PORT", 9092),
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
