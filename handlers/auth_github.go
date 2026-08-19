package handlers

import (
	"encoding/json"
	"forum/config"
	"forum/db"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// GitHubLoginHandler begins the OAuth flow by redirecting the user to GitHub's authorization page.
func GitHubLoginHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Build GitHub OAuth authorization URL
		params := url.Values{}
		params.Add("client_id", config.GitHubClientID)
		params.Add("redirect_uri", config.RedirectURL("/auth/github/callback"))
		params.Add("scope", "user:email") // required to access verified email addresses

		authURL := "https://github.com/login/oauth/authorize?" + params.Encode()

		// Redirect user to GitHub OAuth
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// GitHubCallbackHandler processes the OAuth callback after GitHub redirects back with a code.
func GitHubCallbackHandler(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Extract authorization code from query parameters
		code := r.URL.Query().Get("code")
		if code == "" {
			RenderError(w, http.StatusBadRequest, "Missing authorization code.")
			return
		}

		// 2. Exchange authorization code for an access token
		clientID := config.GitHubClientID
		clientSecret := config.GitHubClientSecret

		tokenReqBody := url.Values{}
		tokenReqBody.Set("client_id", clientID)
		tokenReqBody.Set("client_secret", clientSecret)
		tokenReqBody.Set("code", code)
		tokenReqBody.Set("redirect_uri", config.RedirectURL("/auth/github/callback"))

		// GitHub requires Accept: application/json for token responses
		req, _ := http.NewRequest(
			"POST",
			"https://github.com/login/oauth/access_token",
			strings.NewReader(tokenReqBody.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		tokenResp, err := http.DefaultClient.Do(req)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to exchange code for token.")
			return
		}
		defer tokenResp.Body.Close()

		// Parse token response
		var tokenData struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			Scope       string `json:"scope"`
		}

		if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse token response.")
			return
		}

		if tokenData.AccessToken == "" {
			RenderError(w, http.StatusUnauthorized, "GitHub did not return an access token.")
			return
		}

		// 3. Retrieve GitHub user profile using the access token
		reqUser, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
		reqUser.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
		reqUser.Header.Set("Accept", "application/vnd.github+json")

		userResp, err := http.DefaultClient.Do(reqUser)
		if err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to fetch GitHub user.")
			return
		}
		defer userResp.Body.Close()

		// Handle non‑200 responses
		if userResp.StatusCode < 200 || userResp.StatusCode >= 300 {
			body, _ := io.ReadAll(userResp.Body)
			log.Println("github user fetch failed:", userResp.StatusCode, string(body))
			RenderError(w, http.StatusInternalServerError, "Failed to fetch GitHub user.")
			return
		}

		// Parse basic GitHub user info
		var ghUser struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
			Email string `json:"email"`
		}

		if err := json.NewDecoder(userResp.Body).Decode(&ghUser); err != nil {
			RenderError(w, http.StatusInternalServerError, "Failed to parse GitHub user.")
			return
		}

		// 4. If email is missing, fetch verified emails from /user/emails
		if ghUser.Email == "" {
			reqEmails, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
			reqEmails.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
			reqEmails.Header.Set("Accept", "application/vnd.github+json")

			emailsResp, err := http.DefaultClient.Do(reqEmails)
			if err == nil {
				defer emailsResp.Body.Close()

				if emailsResp.StatusCode >= 200 && emailsResp.StatusCode < 300 {
					var emails []struct {
						Email    string `json:"email"`
						Primary  bool   `json:"primary"`
						Verified bool   `json:"verified"`
					}

					// Parse email list
					if err := json.NewDecoder(emailsResp.Body).Decode(&emails); err == nil {

						// Choose best email: primary+verified → verified → any
						chosen := ""
						for _, e := range emails {
							if e.Primary && e.Verified {
								chosen = e.Email
								break
							}
						}
						if chosen == "" {
							for _, e := range emails {
								if e.Verified {
									chosen = e.Email
									break
								}
							}
						}
						if chosen == "" && len(emails) > 0 {
							chosen = emails[0].Email
						}

						ghUser.Email = chosen
					}
				} else {
					body, _ := io.ReadAll(emailsResp.Body)
					log.Println("github emails fetch failed:", emailsResp.StatusCode, string(body))
				}
			} else {
				log.Println("github emails request failed:", err)
			}
		}

		// 5. Create or link the OAuth user, then establish a session
		provider := "github"
		oauthID := strconv.FormatInt(ghUser.ID, 10)

		user, err := CreateOrLinkOAuthUser(store, provider, oauthID, ghUser.Email, ghUser.Login)
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

		// 6. Redirect user to home page after successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
