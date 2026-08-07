package config

import "os"

type Config struct {
	ServerPort  string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	RedisAddr   string
	RedisPass   string
	JWTSecret   string
	SMSAppKey   string
	OSSEndpoint string
	OSSBucket   string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "3306"),
		DBUser:      getEnv("DB_USER", "root"),
		DBPassword:  getEnv("DB_PASSWORD", "root"),
		DBName:      getEnv("DB_NAME", "eqs"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASS", ""),
		JWTSecret:   getEnv("JWT_SECRET", "eqs-secret-key"),
		SMSAppKey:   getEnv("SMS_APP_KEY", ""),
		OSSEndpoint: getEnv("OSS_ENDPOINT", ""),
		OSSBucket:   getEnv("OSS_BUCKET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
