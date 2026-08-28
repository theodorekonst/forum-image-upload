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

func normalizeDiscordUsername(username, discriminator string) string {
	if username == "" {
		username = "discord_user"
	}
	if discriminator != "" && discriminator != "0" {
		return username + "#" + discriminator
	}
	return username
}

// DiscordLoginHandler begins the OAuth flow by redirecting the user to Discord's authorization page.
func DiscordLoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Generate OAuth state
		state, err := generateOAuthState()
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to start Discord login.")
			return
		}
		setOAuthStateCookie(w, state)

		// 2. Build Discord authorization URL
		authBase := "https://discord.com/api/oauth2/authorize"
		params := url.Values{}
		params.Add("client_id", config.DiscordClientID)
		params.Add("redirect_uri", config.RedirectURL("/auth/discord/callback"))
		params.Add("response_type", "code")
		params.Add("scope", "identify email")
		params.Add("state", state)

		authURL := authBase + "?" + params.Encode()

		// 3. Redirect user to Discord OAuth page
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// DiscordCallbackHandler processes the OAuth callback after Discord redirects back with a code.
func DiscordCallbackHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Read authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			RenderError(w, http.StatusBadRequest, "Missing authorization code.")
			return
		}

		// 2. Validate OAuth state
		state := r.URL.Query().Get("state")
		if !validateOAuthState(r, state) {
			RenderError(w, http.StatusBadRequest, "Invalid OAuth state.")
			return
		}
		clearOAuthStateCookie(w)

		// 3. Exchange code for access token
		tokenURL := "https://discord.com/api/oauth2/token"

		tokenReqBody := url.Values{}
		tokenReqBody.Set("client_id", config.DiscordClientID)
		tokenReqBody.Set("client_secret", config.DiscordClientSecret)
		tokenReqBody.Set("grant_type", "authorization_code")
		tokenReqBody.Set("code", code)
		tokenReqBody.Set("redirect_uri", config.RedirectURL("/auth/discord/callback"))

		req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(tokenReqBody.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		tokenResp, err := oauthHTTPClient.Do(req)
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

		// 4. Parse token response
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

		// 5. Retrieve Discord user profile
		userInfoURL := "https://discord.com/api/users/@me"

		reqUser, _ := http.NewRequest("GET", userInfoURL, nil)
		reqUser.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
		reqUser.Header.Set("Accept", "application/json")

		userResp, err := oauthHTTPClient.Do(reqUser)
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

		// 6. Parse user profile
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

		// 7. Create or link OAuth user
		provider := "discord"
		oauthID := dUser.ID
		uname := normalizeDiscordUsername(dUser.Username, dUser.Discriminator)

		user, err := CreateOrLinkOAuthUser(store, provider, oauthID, dUser.Email, uname)
		if err != nil {
			log.Println("error creating or linking oauth user:", err)
			RenderError(w, http.StatusInternalServerError, "Failed to create or link OAuth user.")
			return
		}

		// 8. Create session and set cookie
		if err := CreateSessionAndSetCookie(w, store, user); err != nil {
			log.Println("error creating session:", err)
			RenderError(w, http.StatusInternalServerError, "Failed to create session.")
			return
		}

		// 9. Redirect user to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
