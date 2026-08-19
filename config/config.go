package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const defaultAppBaseURL = "http://localhost:8080"

var (
	AppBaseURL = defaultAppBaseURL

	// GitHub
	GitHubClientID     = ""
	GitHubClientSecret = ""

	// Google
	GoogleClientID     = ""
	GoogleClientSecret = ""

	// Discord
	DiscordClientID     = ""
	DiscordClientSecret = ""
)

func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func validateBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("URL must include scheme and host")
	}
	return nil
}

func RedirectURL(path string) string {
	return strings.TrimRight(AppBaseURL, "/") + path
}

func Load() error {
	AppBaseURL = getenvDefault("APP_BASE_URL", defaultAppBaseURL)
	GitHubClientID = strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID"))
	GitHubClientSecret = strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET"))
	GoogleClientID = strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	GoogleClientSecret = strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	DiscordClientID = strings.TrimSpace(os.Getenv("DISCORD_CLIENT_ID"))
	DiscordClientSecret = strings.TrimSpace(os.Getenv("DISCORD_CLIENT_SECRET"))

	if err := validateBaseURL(AppBaseURL); err != nil {
		return fmt.Errorf("invalid APP_BASE_URL: %w", err)
	}
	if GitHubClientID == "" || GitHubClientSecret == "" {
		return fmt.Errorf("missing GitHub OAuth environment variables")
	}
	if GoogleClientID == "" || GoogleClientSecret == "" {
		return fmt.Errorf("missing Google OAuth environment variables")
	}
	if DiscordClientID == "" || DiscordClientSecret == "" {
		return fmt.Errorf("missing Discord OAuth environment variables")
	}
	return nil
}
