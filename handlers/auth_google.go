package handlers

import (
    "encoding/json"
    "errors"
    "forum/config"
    "forum/db"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/google/uuid"
)

// GoogleLoginHandler starts the OAuth flow by redirecting the user to Google.
func GoogleLoginHandler(store *db.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        params := url.Values{}
        params.Add("client_id", config.GoogleClientID)
        params.Add("redirect_uri", "http://localhost:8080/auth/google/callback")
        params.Add("response_type", "code")
        params.Add("scope", "email profile")
        params.Add("access_type", "offline")
        params.Add("prompt", "consent")

        authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()

        http.Redirect(w, r, authURL, http.StatusFound)
    }
}

// GoogleCallbackHandler handles the OAuth callback from Google.
func GoogleCallbackHandler(store *db.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        // 1. Read ?code from query
        code := r.URL.Query().Get("code")
        if code == "" {
            RenderError(w, http.StatusBadRequest, "Missing authorization code.")
            return
        }

        // 2. Exchange code for access token
        tokenReqBody := url.Values{}
        tokenReqBody.Set("client_id", config.GoogleClientID)
        tokenReqBody.Set("client_secret", config.GoogleClientSecret)
        tokenReqBody.Set("code", code)
        tokenReqBody.Set("grant_type", "authorization_code")
        tokenReqBody.Set("redirect_uri", "http://localhost:8080/auth/google/callback")

        req, _ := http.NewRequest(
            "POST",
            "https://oauth2.googleapis.com/token",
            strings.NewReader(tokenReqBody.Encode()),
        )
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

        tokenResp, err := http.DefaultClient.Do(req)
        if err != nil {
            RenderError(w, http.StatusInternalServerError, "Failed to exchange code for token.")
            return
        }
        defer tokenResp.Body.Close()

        var tokenData struct {
            AccessToken string `json:"access_token"`
            IDToken     string `json:"id_token"`
            TokenType   string `json:"token_type"`
            ExpiresIn   int    `json:"expires_in"`
        }

        if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
            RenderError(w, http.StatusInternalServerError, "Failed to parse token response.")
            return
        }

        if tokenData.AccessToken == "" {
            RenderError(w, http.StatusUnauthorized, "Google did not return an access token.")
            return
        }

        // 3. Fetch Google user info
        reqUser, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
        reqUser.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

        userResp, err := http.DefaultClient.Do(reqUser)
        if err != nil {
            RenderError(w, http.StatusInternalServerError, "Failed to fetch Google user.")
            return
        }
        defer userResp.Body.Close()

        var gUser struct {
            ID    string `json:"id"`
            Email string `json:"email"`
            Name  string `json:"name"`
        }

        if err := json.NewDecoder(userResp.Body).Decode(&gUser); err != nil {
            RenderError(w, http.StatusInternalServerError, "Failed to parse Google user.")
            return
        }

        // 4. Check if user exists or create new OAuth user
        provider := "google"
        oauthID := gUser.ID

        user, err := store.GetUserByOAuth(provider, oauthID)
        if errors.Is(err, db.ErrNotFound) {
            userID, err := store.CreateOAuthUser(gUser.Email, gUser.Name, provider, oauthID)
            if err != nil {
                RenderError(w, http.StatusInternalServerError, "Failed to create OAuth user.")
                return
            }
            user, _ = store.GetUserByID(userID)
        }

        // 5. Create session and set cookie
        token := uuid.New().String()
        expiresAt := time.Now().Add(24 * time.Hour)

        _, err = store.CreateSession(user.ID, token, expiresAt)
        if err != nil {
            RenderError(w, http.StatusInternalServerError, "Failed to create session.")
            return
        }

        http.SetCookie(w, &http.Cookie{
            Name:     "session",
            Value:    token,
            Path:     "/",
            Expires:  expiresAt,
            HttpOnly: true,
        })

        // 6. Redirect to home
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}
