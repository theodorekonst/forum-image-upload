package handlers

import (
	"encoding/json"
	"forum/config"
	"forum/db"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// DiscordLoginHandler begins the OAuth flow by redirecting the user to Discord's authorization page.
func DiscordLoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Determine authorization endpoint (fallback to default if not configured)
		authBase := config.DiscordAuthURL
		if authBase == "" {
			authBase = "https://discord.com/api/oauth2/authorize"
		}

		// Build Discord OAuth authorization URL
		params := url.Values{}
		params.Add("client_id", config.DiscordClientID)
		params.Add("redirect_uri", config.DiscordRedirectURL)
		params.Add("response_type", "code")
		params.Add("scope", "identify email") // request basic profile + email

		authURL := authBase + "?" + params.Encode()

		// Redirect user to Discord OAuth
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// DiscordCallbackHandler processes the OAuth callback after Discord redirects back with a code.
func DiscordCallbackHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Extract authorization code from query parameters
		code := r.URL.Query().Get("code")
		if code == "" {
			RenderError(w, http.StatusBadRequest, "Missing authorization code.")
			return
		}

		// 2. Exchange authorization code for an access token
		tokenURL := config.DiscordTokenURL
		if tokenURL == "" {
			tokenURL = "https://discord.com/api/oauth2/token"
		}

		tokenReqBody := url.Values{}
		tokenReqBody.Set("client_id", config.DiscordClientID)
		tokenReqBody.Set("client_secret", config.DiscordClientSecret)
		tokenReqBody.Set("grant_type", "authorization_code")
		tokenReqBody.Set("code", code)
		tokenReqBody.Set("redirect_uri", config.DiscordRedirectURL)

		req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(tokenReqBody.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		tokenResp, err := http.DefaultClient.Do(req)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to exchange code for token.")
			return
		}
		defer tokenResp.Body.Close()

		// Handle non‑200 responses
		if tokenResp.StatusCode < 200 || tokenResp.StatusCode >= 300 {
			body, _ := io.ReadAll(tokenResp.Body)
			log.Println("discord token exchange failed:", tokenResp.StatusCode, string(body))
			RenderError(w, http.StatusInternalServerError, "Failed to exchange code for token.")
			return
		}

		// Parse token response
		var tokenData struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			Scope       string `json:"scope"`
		}

		if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse token response.")
			return
		}

		if tokenData.AccessToken == "" {
			RenderError(w, http.StatusUnauthorized, "Discord did not return an access token.")
			return
		}

		// 3. Retrieve Discord user profile using the access token
		userInfoURL := config.DiscordUserInfoURL
		if userInfoURL == "" {
			userInfoURL = "https://discord.com/api/users/@me"
		}

		reqUser, _ := http.NewRequest("GET", userInfoURL, nil)
		reqUser.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
		reqUser.Header.Set("Accept", "application/json")

		userResp, err := http.DefaultClient.Do(reqUser)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to fetch Discord user.")
			return
		}
		defer userResp.Body.Close()

		// Handle non‑200 responses
		if userResp.StatusCode < 200 || userResp.StatusCode >= 300 {
			body, _ := io.ReadAll(userResp.Body)
			log.Println("discord user fetch failed:", userResp.StatusCode, string(body))
			RenderError(w, http.StatusInternalServerError, "Failed to fetch Discord user.")
			return
		}

		// Parse user profile
		var dUser struct {
			ID            string `json:"id"`
			Username      string `json:"username"`
			Discriminator string `json:"discriminator"`
			Email         string `json:"email"`
		}

		if err := json.NewDecoder(userResp.Body).Decode(&dUser); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse Discord user.")
			return
		}

		// 4. Create or link the OAuth user, then establish a session
		provider := "discord"
		oauthID := dUser.ID

		// Build a display username (fallback if missing)
		uname := dUser.Username
		if uname == "" {
			uname = "discord_user"
		}
		if dUser.Discriminator != "" {
			uname = uname + "#" + dUser.Discriminator
		}

		user, err := CreateOrLinkOAuthUser(store, provider, oauthID, dUser.Email, uname)
		if err != nil {
			log.Println("error creating or linking oauth user:", err)
			RenderError(w, http.StatusInternalServerError, "Failed to create or link OAuth user.")
			return
		}

		// Create session and set cookie
		if err := CreateSessionAndSetCookie(w, store, user); err != nil {
			log.Println("error creating session:", err)
			RenderError(w, http.StatusInternalServerError, "Failed to create session.")
			return
		}

		// 5. Redirect user to home page after successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
