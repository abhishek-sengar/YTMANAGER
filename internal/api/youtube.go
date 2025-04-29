package api

import (
	"context"
	_ "fmt"
	"net/http"
	"os"

	"github.com/abhishek-sengar/ytmanager/internal/db"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
		},
		Endpoint: google.Endpoint,
	}
}

// YoutubeAuth redirects Owner to Google's OAuth consent screen
func YoutubeAuth(c *gin.Context) {
	cfg := getGoogleOauthConfig()
	authURL := cfg.AuthCodeURL("randomstate")
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// YoutubeCallback handles OAuth callback, exchanges code for tokens
// YoutubeCallback
func YoutubeCallback(c *gin.Context) {
	cfg := getGoogleOauthConfig()

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	// Get the logged-in user from token (we assume Owner is logged-in)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(string)

	// Save tokens into database for the Owner
	_, err = db.DB.Exec(`
		UPDATE users
		SET youtube_access_token = $1,
		    youtube_refresh_token = $2
		WHERE id = $3
	`, token.AccessToken, token.RefreshToken, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save tokens: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "YouTube account connected successfully!"})
}
