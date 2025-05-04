package api

import (
	"fmt"
	"github.com/abhishek-sengar/ytmanager/internal/db"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetEditorWorkspaces returns all owners and their channels available to the editor
func GetEditorWorkspaces(c *gin.Context) {
	editorIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	editorID := editorIDInterface.(string)

	query := `
		SELECT
			u.id as owner_id,
			u.email as owner_name,
			c.id as channel_id,
			c.name as channel_name
		FROM
			workspaces w
			JOIN workspace_owners wo ON w.id = wo.workspace_id
			JOIN users u ON wo.owner_id = u.id
			JOIN workspace_channels wc ON w.id = wc.workspace_id
			JOIN channels c ON wc.channel_id = c.id
		WHERE
			w.editor_id = $1
	`

	rows, err := db.DB.Query(query, editorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed: " + err.Error()})
		return
	}
	defer rows.Close()

	type Channel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	type Owner struct {
		OwnerID   string    `json:"owner_id"`
		OwnerName string    `json:"owner_name"`
		Channels  []Channel `json:"channels"`
	}

	ownersMap := make(map[string]*Owner)

	for rows.Next() {
		var ownerID, ownerName, channelID, channelName string
		if err := rows.Scan(&ownerID, &ownerName, &channelID, &channelName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Row scan error: " + err.Error()})
			return
		}

		if _, exists := ownersMap[ownerID]; !exists {
			ownersMap[ownerID] = &Owner{
				OwnerID:   ownerID,
				OwnerName: ownerName,
				Channels:  []Channel{},
			}
		}

		ownersMap[ownerID].Channels = append(ownersMap[ownerID].Channels, Channel{
			ID:   channelID,
			Name: channelName,
		})
	}

	var owners []Owner
	for _, o := range ownersMap {
		owners = append(owners, *o)
	}

	c.JSON(http.StatusOK, owners)
}

// ProjectResponse defines the structure of returned video cards
type ProjectResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ChannelName string `json:"channel_name"`
	OwnerName   string `json:"owner_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func GetEditorRecentProjects(c *gin.Context) {
	editorIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	editorID := editorIDInterface.(string)

	ownerID := c.Query("owner_id")
	channelID := c.Query("channel_id")

	query := `
		SELECT
			p.id, p.title, p.description, p.status,
			ch.name AS channel_name,
			u.email AS owner_name,
			p.created_at, p.updated_at
		FROM projects p
		JOIN channels ch ON p.channel_id = ch.id
		JOIN users u ON p.owner_id = u.id
		WHERE p.editor_id = $1
	`

	args := []interface{}{editorID}
	argIndex := 2

	if channelID != "" {
		query += " AND p.channel_id = $" + itoa(argIndex)
		args = append(args, channelID)
		argIndex++
	} else if ownerID != "" {
		query += " AND p.owner_id = $" + itoa(argIndex)
		args = append(args, ownerID)
		argIndex++
	}

	query += " ORDER BY p.updated_at DESC LIMIT 20"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query projects: " + err.Error()})
		return
	}
	defer rows.Close()

	var projects []ProjectResponse
	for rows.Next() {
		var p ProjectResponse
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Status,
			&p.ChannelName, &p.OwnerName,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Row scan error: " + err.Error()})
			return
		}
		projects = append(projects, p)
	}

	c.JSON(http.StatusOK, projects)
}

// helper to format placeholders like $2, $3 in query
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
