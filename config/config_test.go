package config

import (
	"os"
	"os/exec"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	if exitCode := runLoadCase(t, map[string]string{
		"GITHUB_CLIENT_ID":      "gh-id",
		"GITHUB_CLIENT_SECRET":  "gh-secret",
		"GOOGLE_CLIENT_ID":      "google-id",
		"GOOGLE_CLIENT_SECRET":  "google-secret",
		"DISCORD_CLIENT_ID":     "disc-id",
		"DISCORD_CLIENT_SECRET": "disc-secret",
		"DISCORD_REDIRECT_URL":  "http://localhost/callback",
		"DISCORD_AUTH_URL":      "https://discord.com/api/oauth2/authorize",
		"DISCORD_TOKEN_URL":     "https://discord.com/api/oauth2/token",
		"DISCORD_USERINFO_URL":  "https://discord.com/api/users/@me",
	}); exitCode != 0 {
		t.Fatalf("Load() should succeed with valid config, got exit code %d", exitCode)
	}
}

func TestLoad_MissingGitHubConfig(t *testing.T) {
	if exitCode := runLoadCase(t, map[string]string{
		"GITHUB_CLIENT_ID":      "",
		"GITHUB_CLIENT_SECRET":  "",
		"GOOGLE_CLIENT_ID":      "google-id",
		"GOOGLE_CLIENT_SECRET":  "google-secret",
		"DISCORD_CLIENT_ID":     "disc-id",
		"DISCORD_CLIENT_SECRET": "disc-secret",
		"DISCORD_REDIRECT_URL":  "http://localhost/callback",
		"DISCORD_AUTH_URL":      "https://discord.com/api/oauth2/authorize",
		"DISCORD_TOKEN_URL":     "https://discord.com/api/oauth2/token",
		"DISCORD_USERINFO_URL":  "https://discord.com/api/users/@me",
	}); exitCode == 0 {
		t.Fatal("Load() should fail when GitHub config is missing")
	}
}

func TestLoad_MissingGoogleConfig(t *testing.T) {
	if exitCode := runLoadCase(t, map[string]string{
		"GITHUB_CLIENT_ID":      "gh-id",
		"GITHUB_CLIENT_SECRET":  "gh-secret",
		"GOOGLE_CLIENT_ID":      "",
		"GOOGLE_CLIENT_SECRET":  "",
		"DISCORD_CLIENT_ID":     "disc-id",
		"DISCORD_CLIENT_SECRET": "disc-secret",
		"DISCORD_REDIRECT_URL":  "http://localhost/callback",
		"DISCORD_AUTH_URL":      "https://discord.com/api/oauth2/authorize",
		"DISCORD_TOKEN_URL":     "https://discord.com/api/oauth2/token",
		"DISCORD_USERINFO_URL":  "https://discord.com/api/users/@me",
	}); exitCode == 0 {
		t.Fatal("Load() should fail when Google config is missing")
	}
}

func TestLoad_MissingDiscordConfig(t *testing.T) {
	if exitCode := runLoadCase(t, map[string]string{
		"GITHUB_CLIENT_ID":      "gh-id",
		"GITHUB_CLIENT_SECRET":  "gh-secret",
		"GOOGLE_CLIENT_ID":      "google-id",
		"GOOGLE_CLIENT_SECRET":  "google-secret",
		"DISCORD_CLIENT_ID":     "",
		"DISCORD_CLIENT_SECRET": "",
		"DISCORD_REDIRECT_URL":  "",
		"DISCORD_AUTH_URL":      "https://discord.com/api/oauth2/authorize",
		"DISCORD_TOKEN_URL":     "https://discord.com/api/oauth2/token",
		"DISCORD_USERINFO_URL":  "https://discord.com/api/users/@me",
	}); exitCode == 0 {
		t.Fatal("Load() should fail when Discord config is missing")
	}
}

func runLoadCase(t *testing.T, values map[string]string) int {
	t.Helper()

	env := os.Environ()
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLoadHelperProcess")
	cmd.Env = append(env, "GO_WANT_HELPER_PROCESS=1")

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

func TestLoadHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	Load()
	os.Exit(0)
}
