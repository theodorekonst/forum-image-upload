package config

import (
	"log"
	"os"
)

var (
	GitHubClientID     string
	GitHubClientSecret string
)

func Load() {
	GitHubClientID = os.Getenv("GITHUB_CLIENT_ID")
	GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	if GitHubClientID == "" || GitHubClientSecret == "" {
		log.Fatal("Missing GitHub OAuth environment variables")
	}
}
