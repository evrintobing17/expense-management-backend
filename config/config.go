package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
	JWTSecret      string
	ServerPort     string
	PaymentAPIURL  string
	WorkerInterval int
}

func Load() *Config {
	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBName:         getEnv("DB_NAME", "expense_db"),
		DBUser:         getEnv("DB_USER", "expense_user"),
		DBPassword:     getEnv("DB_PASSWORD", "expense_password"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		PaymentAPIURL:  getEnv("PAYMENT_API_URL", "https://1620e98f-7759-431c-a2aa-f449d591150b.mock.pstmn.io"),
		WorkerInterval: getEnvAsInt("WORKER_INTERVAL", 30),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
