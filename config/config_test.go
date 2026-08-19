package config

import "testing"

func TestLoad_Success(t *testing.T) {
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
	t.Setenv("GITHUB_CLIENT_ID", "gh-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "gh-secret")
	t.Setenv("GOOGLE_CLIENT_ID", "google-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("DISCORD_CLIENT_ID", "disc-id")
	t.Setenv("DISCORD_CLIENT_SECRET", "disc-secret")

	if err := Load(); err != nil {
		t.Fatalf("Load() should succeed with valid config: %v", err)
	}
}

func TestLoad_MissingGitHubConfig(t *testing.T) {
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
	t.Setenv("GITHUB_CLIENT_ID", "")
	t.Setenv("GITHUB_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_CLIENT_ID", "google-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("DISCORD_CLIENT_ID", "disc-id")
	t.Setenv("DISCORD_CLIENT_SECRET", "disc-secret")

	if err := Load(); err == nil {
		t.Fatal("Load() should fail when GitHub config is missing")
	}
}

func TestLoad_MissingGoogleConfig(t *testing.T) {
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
	t.Setenv("GITHUB_CLIENT_ID", "gh-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "gh-secret")
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("DISCORD_CLIENT_ID", "disc-id")
	t.Setenv("DISCORD_CLIENT_SECRET", "disc-secret")

	if err := Load(); err == nil {
		t.Fatal("Load() should fail when Google config is missing")
	}
}

func TestLoad_MissingDiscordConfig(t *testing.T) {
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
	t.Setenv("GITHUB_CLIENT_ID", "gh-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "gh-secret")
	t.Setenv("GOOGLE_CLIENT_ID", "google-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("DISCORD_CLIENT_ID", "")
	t.Setenv("DISCORD_CLIENT_SECRET", "")

	if err := Load(); err == nil {
		t.Fatal("Load() should fail when Discord config is missing")
	}
}

func TestLoad_InvalidAppBaseURL(t *testing.T) {
	t.Setenv("APP_BASE_URL", "localhost:8080")
	t.Setenv("GITHUB_CLIENT_ID", "gh-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "gh-secret")
	t.Setenv("GOOGLE_CLIENT_ID", "google-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	t.Setenv("DISCORD_CLIENT_ID", "disc-id")
	t.Setenv("DISCORD_CLIENT_SECRET", "disc-secret")

	if err := Load(); err == nil {
		t.Fatal("Load() should fail when APP_BASE_URL is invalid")
	}
}

func TestRedirectURL(t *testing.T) {
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
	if err := Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if got, want := RedirectURL("/auth/github/callback"), "http://localhost:8080/auth/github/callback"; got != want {
		t.Fatalf("RedirectURL() = %q, want %q", got, want)
	}
}
