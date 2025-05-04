package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"errors"

	"github.com/abhishek-sengar/ytmanager/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Setup OAuth2 config
// var googleOauthConfig = &oauth2.Config{
// 	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
// 	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
// 	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
// 	Scopes: []string{
// 		youtube.YoutubeUploadScope,
// 		youtube.YoutubeScope,
// 	},
// 	Endpoint: google.Endpoint,
// }

func getGoogleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			youtube.YoutubeUploadScope,
			youtube.YoutubeScope,
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

// Helper to parse userID from JWT
func ParseUserIDFromJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	// Check for 'sub' claim first since that's what Login uses
	userID, ok := claims["sub"].(string)
	if !ok {
		// Fall back to 'user_id' claim if 'sub' not found
		userID, ok = claims["user_id"].(string)
		if !ok {
			return "", errors.New("user_id not found in token")
		}
	}
	return userID, nil
}

// YoutubeAuth redirects Owner to Google's OAuth consent screen
func YoutubeAuth(c *gin.Context) {
	cfg := getGoogleOauthConfig()
	state := c.Query("state") // get from frontend
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing state (JWT)"})
		return
	}
	authURL := cfg.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// YoutubeCallback handles OAuth callback, exchanges code for tokens
func YoutubeCallback(c *gin.Context) {
	cfg := getGoogleOauthConfig()

	code := c.Query("code")
	state := c.Query("state") // this is the JWT

	fmt.Printf("Received callback with code: %s\n", code)
	fmt.Printf("Received state: %s\n", state)

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code or state not found"})
		return
	}

	// Validate the JWT (state) and extract userID
	userID, err := ParseUserIDFromJWT(state)
	if err != nil {
		fmt.Printf("Error parsing JWT: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	fmt.Printf("Successfully extracted userID: %s\n", userID)

	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("Error exchanging code for token: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}
	fmt.Printf("Successfully exchanged code for token\n")

	// Fetch email from Google UserInfo API
	client := cfg.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Error fetching user info: %v, Status: %d\n", err, resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}
	defer resp.Body.Close()
	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		fmt.Printf("Error decoding user info: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}
	fmt.Printf("Successfully fetched user email: %s\n", userInfo.Email)

	// Save tokens into database for the Owner
	_, err = db.DB.Exec(`
		INSERT INTO youtube_accounts (user_id, email, access_token, refresh_token)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, email) DO UPDATE
		SET access_token = $3, refresh_token = $4, updated_at = now()
	`, userID, userInfo.Email, token.AccessToken, token.RefreshToken)

	if err != nil {
		fmt.Printf("Error saving tokens to database: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tokens: " + err.Error()})
		return
	}
	fmt.Printf("Successfully saved tokens to database\n")

	// Fetch channels for this account
	ytService, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		fmt.Printf("Error creating YouTube service: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create YouTube service"})
		return
	}

	call := ytService.Channels.List([]string{"snippet"}).Mine(true)
	channelsResp, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching channels: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
		return
	}

	var channels []map[string]interface{}
	for _, ch := range channelsResp.Items {
		channels = append(channels, map[string]interface{}{
			"id":      ch.Id,
			"name":    ch.Snippet.Title,
			"iconUrl": ch.Snippet.Thumbnails.Default.Url,
		})
	}

	// Encode channels as JSON for URL parameter
	channelsJSON, err := json.Marshal(channels)
	if err != nil {
		fmt.Printf("Error encoding channels: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode channels"})
		return
	}

	// Redirect back to frontend with success status and channels
	redirectURL := fmt.Sprintf(
		"http://localhost:5173/oauth-callback?status=success&email=%s&user_id=%s&channels=%s",
		url.QueryEscape(userInfo.Email),
		url.QueryEscape(userID),
		url.QueryEscape(string(channelsJSON)),
	)
	fmt.Printf("Redirecting to: %s\n", redirectURL)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GET /api/youtube/unattached-channels
func GetUnattachedChannels(c *gin.Context) {
	userID := c.GetString("userID")
	fmt.Printf("Fetching unattached channels for user: %s\n", userID)

	// 1. Fetch all YouTube accounts for this user
	rows, err := db.DB.Query(`SELECT id, access_token, refresh_token, email FROM youtube_accounts WHERE user_id = $1`, userID)
	if err != nil {
		fmt.Printf("Error fetching YouTube accounts: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to fetch YouTube accounts"})
		return
	}
	defer rows.Close()

	var allChannels []map[string]interface{}
	attachedChannelIDs := map[string]bool{}

	// 2. Get all channel IDs already attached to dashboard
	rows2, err := db.DB.Query(`SELECT yt_channel_id FROM channels WHERE owner_id = $1`, userID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var cid string
			rows2.Scan(&cid)
			attachedChannelIDs[cid] = true
		}
	}

	// 3. For each YouTube account, fetch channels
	for rows.Next() {
		var accID, accessToken, refreshToken, email string
		rows.Scan(&accID, &accessToken, &refreshToken, &email)
		fmt.Printf("Processing YouTube account: %s\n", email)
		fmt.Printf("Access token length: %d, Refresh token length: %d\n", len(accessToken), len(refreshToken))

		cfg := &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes:       []string{youtube.YoutubeReadonlyScope},
			Endpoint:     google.Endpoint,
		}
		token := &oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken}
		client := cfg.Client(context.Background(), token)
		ytService, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
		if err != nil {
			fmt.Printf("Error creating YouTube service for account %s: %v\n", email, err)
			continue
		}
		call := ytService.Channels.List([]string{"snippet"}).Mine(true)
		resp, err := call.Do()
		if err != nil {
			fmt.Printf("Error fetching channels for account %s: %v\n", email, err)
			// Try to get more details about the error
			if youtubeErr, ok := err.(*googleapi.Error); ok {
				fmt.Printf("YouTube API Error: Code=%d, Message=%s\n", youtubeErr.Code, youtubeErr.Message)
			}
			continue
		}
		fmt.Printf("Found %d channels for account %s\n", len(resp.Items), email)
		for _, ch := range resp.Items {
			if attachedChannelIDs[ch.Id] {
				fmt.Printf("Skipping already attached channel: %s (ID: %s)\n", ch.Snippet.Title, ch.Id)
				continue // already attached
			}
			fmt.Printf("Adding channel: %s (ID: %s)\n", ch.Snippet.Title, ch.Id)
			allChannels = append(allChannels, map[string]interface{}{
				"id":                 ch.Id,
				"name":               ch.Snippet.Title,
				"iconUrl":            ch.Snippet.Thumbnails.Default.Url,
				"email":              email,
				"youtube_account_id": accID,
			})
		}
	}

	fmt.Printf("Returning %d unattached channels\n", len(allChannels))
	c.JSON(200, gin.H{"channels": allChannels})
}

// POST /api/youtube/add-channels
func AddChannelsToDashboard(c *gin.Context) {
	userID := c.GetString("userID")
	fmt.Printf("Updating channels for user: %s\n", userID)

	var req struct {
		Channels []struct {
			ID               string `json:"id"`
			Name             string `json:"name"`
			IconUrl          string `json:"iconUrl"`
			Email            string `json:"email"`
			YouTubeAccountID string `json:"youtube_account_id"`
		} `json:"channels"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Start a transaction
	tx, err := db.DB.Begin()
	if err != nil {
		fmt.Printf("Error starting transaction: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get current channels for this user
	rows, err := tx.Query(`SELECT yt_channel_id FROM channels WHERE owner_id = $1`, userID)
	if err != nil {
		fmt.Printf("Error fetching current channels: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to fetch current channels"})
		return
	}
	defer rows.Close()

	currentChannels := make(map[string]bool)
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			fmt.Printf("Error scanning channel ID: %v\n", err)
			c.JSON(500, gin.H{"error": "Failed to scan channel ID"})
			return
		}
		currentChannels[channelID] = true
	}

	// Process channels to add
	channelsToAdd := make(map[string]bool)
	for _, ch := range req.Channels {
		channelsToAdd[ch.ID] = true
		if !currentChannels[ch.ID] {
			// Add new channel
			_, err := tx.Exec(`
				INSERT INTO channels (owner_id, youtube_account_id, yt_channel_id, name, icon_url, email)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, userID, ch.YouTubeAccountID, ch.ID, ch.Name, ch.IconUrl, ch.Email)
			if err != nil {
				fmt.Printf("Error adding channel %s: %v\n", ch.Name, err)
				c.JSON(500, gin.H{"error": "Failed to add channel: " + err.Error()})
				return
			}
			fmt.Printf("Added channel: %s\n", ch.Name)
		}
	}

	// Remove channels that are no longer selected
	for channelID := range currentChannels {
		if !channelsToAdd[channelID] {
			_, err := tx.Exec(`DELETE FROM channels WHERE owner_id = $1 AND yt_channel_id = $2`, userID, channelID)
			if err != nil {
				fmt.Printf("Error removing channel %s: %v\n", channelID, err)
				c.JSON(500, gin.H{"error": "Failed to remove channel: " + err.Error()})
				return
			}
			fmt.Printf("Removed channel: %s\n", channelID)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to commit changes"})
		return
	}

	fmt.Printf("Successfully updated channels\n")
	c.JSON(200, gin.H{"message": "Channels updated successfully"})
}
