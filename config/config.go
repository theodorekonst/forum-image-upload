package config

import (
	"log"
	"os"
	"strings"
)

var (
	AppBaseURL = getenvDefault("APP_BASE_URL", "http://localhost:8080")

	// GitHub
	GitHubClientID     = os.Getenv("GITHUB_CLIENT_ID")
	GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	// Google
	GoogleClientID     = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")

	// Discord
	DiscordClientID     = os.Getenv("DISCORD_CLIENT_ID")
	DiscordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
)

func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func RedirectURL(path string) string {
	return strings.TrimRight(AppBaseURL, "/") + path
}

func Load() {
	if GitHubClientID == "" || GitHubClientSecret == "" {
		log.Fatal("Missing GitHub OAuth environment variables")
	}

	if GoogleClientID == "" || GoogleClientSecret == "" {
		log.Fatal("Missing Google OAuth environment variables")
	}

	if DiscordClientID == "" || DiscordClientSecret == "" {
		log.Fatal("Missing Discord OAuth environment variables")
	}
}
