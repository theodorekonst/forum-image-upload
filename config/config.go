package config

import (
	"log"
	"os"
)

var (
	// GitHub
	GitHubClientID     = os.Getenv("GITHUB_CLIENT_ID")
	GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	// Google
	GoogleClientID     = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")

	// Discord
	DiscordClientID     = os.Getenv("DISCORD_CLIENT_ID")
	DiscordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	DiscordRedirectURL  = os.Getenv("DISCORD_REDIRECT_URL")
	DiscordAuthURL      = os.Getenv("DISCORD_AUTH_URL")
	DiscordTokenURL     = os.Getenv("DISCORD_TOKEN_URL")
	DiscordUserInfoURL  = os.Getenv("DISCORD_USERINFO_URL")
)

func Load() {
	// GitHub required
	if GitHubClientID == "" || GitHubClientSecret == "" {
		log.Fatal("Missing GitHub OAuth environment variables")
	}

	// Google required
	if GoogleClientID == "" || GoogleClientSecret == "" {
		log.Fatal("Missing Google OAuth environment variables")
	}

	// Discord required
	if DiscordClientID == "" || DiscordClientSecret == "" || DiscordRedirectURL == "" {
		log.Fatal("Missing Discord OAuth environment variables")
	}
}
