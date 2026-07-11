package commonconfig

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ProjectID              string
	HashSecret             string
	GeminiApiKey           string
	GoogleAuthClientID     string
	GoogleAuthClientSecret string
	GoogleAuthRedirectURL  string
}

func NewConfig(envPath string) (Config, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return Config{}, fmt.Errorf("unable to load env: %w", err)
	}
	return Config{
		ProjectID:              os.Getenv("PROJECT_ID"),
		HashSecret:             os.Getenv("HASH_SECRET"),
		GeminiApiKey:           os.Getenv("GEMINI_API_KEY"),
		GoogleAuthClientID:     os.Getenv("GOOGLE_AUTH_CLIENT_ID"),
		GoogleAuthClientSecret: os.Getenv("GOOGLE_AUTH_CLIENT_SECRET"),
		GoogleAuthRedirectURL:  os.Getenv("GOOGLE_AUTH_REDIRECT_URL"),
	}, nil
}
