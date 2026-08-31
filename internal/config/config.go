package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort             string
	ServerMode             string
	DatabaseURL            string
	RedisURL               string
	RedisCacheTTLHours     int
	NominatimBaseURL       string
	OSRMBaseURL            string
	TollBufferMeters       float64
	TollDeduplicationDist  float64
	CSVFilePath            string
}

func Load() *Config {
	// Load .env file if it exists, ignore error if not found (e.g. In Docker/Prod)
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded, using environment variables")
	}

	return &Config{
		ServerPort:             getEnv("SERVER_PORT", "8080"),
		ServerMode:             getEnv("SERVER_MODE", "debug"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/toll_plaza_db?sslmode=disable"),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6379/0"),
		RedisCacheTTLHours:     getEnvAsInt("REDIS_CACHE_TTL_HOURS", 24),
		NominatimBaseURL:       getEnv("NOMINATIM_BASE_URL", "https://nominatim.openstreetmap.org"),
		OSRMBaseURL:            getEnv("OSRM_BASE_URL", "https://router.project-osrm.org"),
		TollBufferMeters:       getEnvAsFloat("TOLL_BUFFER_METERS", 500.0),
		TollDeduplicationDist:  getEnvAsFloat("TOLL_DEDUPLICATION_METERS", 1000.0),
		CSVFilePath:            getEnv("CSV_FILE_PATH", "./toll_plaza_india.csv"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}

func getEnvAsFloat(key string, fallback float64) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return fallback
	}
	return val
}
