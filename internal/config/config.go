package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisAddr         string
	RedisApiKey       string
	RedisUser         string
	RedisPassword     string
	JWTSecret         string
	JWTTTL            int
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPass          string
	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioPhoneNumber string

	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDBName   string
}

func LoadEnv() *Config {
	// Load .env if present
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, relying on environment variables")
	}

	port, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		log.Fatalf("Invalid SMTP_PORT: %v", err)
	}

	jwtTTL, err := strconv.Atoi(getEnv("JWT_TTL", "3600"))
	if err != nil {
		log.Fatalf("Invalid JWT_TTL: %v", err)
	}

	cfg := &Config{
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisApiKey:       getEnv("REDIS_API_KEY", ""),
		RedisUser:         getEnv("REDIS_USERNAME", "default"),
		RedisPassword:     getEnv("REDIS_PASSWORD", "default"),
		JWTSecret:         getEnv("JWT_SECRET", "supersecret"),
		JWTTTL:            jwtTTL,
		SMTPHost:          getEnv("SMTP_HOST", ""),
		SMTPPort:          port,
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPass:          getEnv("SMTP_PASS", ""),
		TwilioAccountSID:  getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:   getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber: getEnv("TWILIO_PHONE_NUMBER", ""),

		PGHost:     getEnv("PG_HOST", "localhost"),
		PGPort:     getEnv("PG_PORT", "5432"),
		PGUser:     getEnv("PG_USER", "postgres"),
		PGPassword: getEnv("PG_PASSWORD", ""),
		PGDBName:   getEnv("PG_DBNAME", "postgres"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
