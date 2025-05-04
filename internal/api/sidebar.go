package api

import (
	"net/http"

	"github.com/abhishek-sengar/ytmanager/internal/db" // ← adjust to your module path
	"github.com/gin-gonic/gin"
)

// SidebarResponse is the unified response shape
type SidebarResponse struct {
	Channels []SidebarItem `json:"channels"`
	Partners []SidebarItem `json:"partners"` // owners for editors, editors for owners
}

type SidebarItem struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	IconURL          string `json:"iconUrl"`
	Email            string `json:"email"`
	YouTubeAccountID string `json:"youtube_account_id"`
}

// GetSidebarData returns channels + partners based on userRole
func GetSidebarData(c *gin.Context) {
	userIDInterface, ok := c.Get("userID")
	roleInterface, rok := c.Get("userRole")
	if !ok || !rok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDInterface.(string)
	role := roleInterface.(string)

	var (
		channels []SidebarItem
		partners []SidebarItem
	)

	switch role {
	case "editor":
		// Channels the editor is attached to
		chRows, err := db.DB.Query(`
			SELECT c.id, c.name, c.icon_url, c.email, c.youtube_account_id
			FROM editors_channels ec
			JOIN channels c ON ec.channel_id = c.id
			WHERE ec.editor_id = $1
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels: " + err.Error()})
			return
		}
		defer chRows.Close()
		for chRows.Next() {
			var itm SidebarItem
			if err := chRows.Scan(&itm.ID, &itm.Name, &itm.IconURL, &itm.Email, &itm.YouTubeAccountID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Channel scan error: " + err.Error()})
				return
			}
			channels = append(channels, itm)
		}

		// Partners: all owners of those channels
		ownerRows, err := db.DB.Query(`
			SELECT DISTINCT u.id, u.name
			FROM editors_channels ec
			JOIN channels c ON ec.channel_id = c.id
			JOIN users u ON c.owner_id = u.id
			WHERE ec.editor_id = $1
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch owners: " + err.Error()})
			return
		}
		defer ownerRows.Close()
		for ownerRows.Next() {
			var itm SidebarItem
			if err := ownerRows.Scan(&itm.ID, &itm.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Owner scan error: " + err.Error()})
				return
			}
			partners = append(partners, itm)
		}

	case "owner":
		// Channels owned by this user
		chRows, err := db.DB.Query(`
			SELECT id, name, icon_url, email, youtube_account_id
			FROM channels
			WHERE owner_id = $1
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels: " + err.Error()})
			return
		}
		defer chRows.Close()
		for chRows.Next() {
			var itm SidebarItem
			if err := chRows.Scan(&itm.ID, &itm.Name, &itm.IconURL, &itm.Email, &itm.YouTubeAccountID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Channel scan error: " + err.Error()})
				return
			}
			channels = append(channels, itm)
		}

		// Partners: all editors attached to any of their channels
		edRows, err := db.DB.Query(`
			SELECT DISTINCT u.id, u.name
			FROM editors_channels ec
			JOIN channels c ON ec.channel_id = c.id
			JOIN users u ON ec.editor_id = u.id
			WHERE c.owner_id = $1
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch editors: " + err.Error()})
			return
		}
		defer edRows.Close()
		for edRows.Next() {
			var itm SidebarItem
			if err := edRows.Scan(&itm.ID, &itm.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Editor scan error: " + err.Error()})
				return
			}
			partners = append(partners, itm)
		}

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Unsupported role"})
		return
	}

	c.JSON(http.StatusOK, SidebarResponse{
		Channels: channels,
		Partners: partners,
	})
}
