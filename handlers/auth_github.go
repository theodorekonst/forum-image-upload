package handlers

import (
	"forum/config"
	"forum/db"
	"net/http"
	"net/url"
)

// GitHubLoginHandler starts the OAuth flow by redirecting the user to GitHub.
func GitHubLoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		params := url.Values{}
		params.Add("client_id", config.GitHubClientID)
		params.Add("redirect_uri", "http://localhost:8080/auth/github/callback")
		params.Add("scope", "user:email")

		authURL := "https://github.com/login/oauth/authorize?" + params.Encode()

		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// GitHubCallbackHandler handles the OAuth callback from GitHub.
func GitHubCallbackHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Read ?code from query
		code := r.URL.Query().Get("code")
		if code == "" {
			RenderError(w, http.StatusBadRequest, "Missing authorization code.")
			return
		}

		// 2. Exchange code for access token
		// (Step 4)

		// 3. Fetch GitHub user info
		// (Step 4)

		// 4. Check if user exists or create new OAuth user
		// (Step 5)

		// 5. Create session and set cookie
		// (Step 6)

		// 6. Redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
