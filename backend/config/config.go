package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURL    string
	MongoDBName string
	RedisURL    string
	JWTSecret   string
	Port        string
}

var AppConfig *Config

// Load reads .env (if present) and populates AppConfig.
// Missing .env is not fatal -- real environments may inject vars directly.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	cfg := &Config{
		MongoURL:    getEnv("MONGO_URL", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "live_polling"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "change_this_secret_in_production"),
		Port:        getEnv("PORT", "8080"),
	}

	if cfg.JWTSecret == "change_this_secret_in_production" {
		log.Println("WARNING: using default JWT_SECRET. Set JWT_SECRET in .env for production.")
	}

	AppConfig = cfg
	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
