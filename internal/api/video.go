package api

import (
	"net/http"
	"time"

	"github.com/abhishek-sengar/ytmanager/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateProjectRequest represents the JSON body for creating a project
type CreateProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	VideoPath   string `json:"video_path" binding:"required"` // for now you give video file path, later we upload files
}

// CreateProject allows an Editor to create a new project
func CreateProject(c *gin.Context) {
	var req CreateProjectRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from context
	editorIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	editorID := editorIDInterface.(string)

	// For now, we won't assign Owner directly. We assume Owner ID will be assigned manually later or a simple logic for now.
	var ownerID string
	err := db.DB.QueryRow(`SELECT id FROM users WHERE role = 'owner' LIMIT 1`).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No owner found to assign project"})
		return
	}

	// Insert new project
	query := `
        INSERT INTO projects (id, title, description, video_path, status, editor_id, owner_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	_, err = db.DB.Exec(
		query,
		uuid.New().String(),
		req.Title,
		req.Description,
		req.VideoPath,
		"pending",
		editorID,
		ownerID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project created successfully"})
}

// Project represents project data
type Project struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	VideoPath   string    `json:"video_path"`
	Status      string    `json:"status"`
	EditorID    string    `json:"editor_id"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetProjects fetches all projects for the logged-in user
func GetProjects(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	userID := userIDInterface.(string)

	// Fetch projects where the user is Editor or Owner
	rows, err := db.DB.Query(`
        SELECT id, title, description, video_path, status, editor_id, owner_id, created_at, updated_at
        FROM projects
        WHERE editor_id = $1 OR owner_id = $1
        ORDER BY created_at DESC
    `, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects: " + err.Error()})
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.VideoPath, &p.Status,
			&p.EditorID, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse project: " + err.Error()})
			return
		}
		projects = append(projects, p)
	}

	c.JSON(http.StatusOK, projects)
}

// AddNoteRequest represents the body to add a note
type AddNoteRequest struct {
	Timestamp int    `json:"timestamp" binding:"required"` // in seconds
	Content   string `json:"content" binding:"required"`   // comment text
}

// AddNote allows Owner to add a note to a project
func AddNote(c *gin.Context) {
	projectID := c.Param("id")
	var req AddNoteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user info
	userRoleInterface, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}
	userRole := userRoleInterface.(string)

	if userRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can add notes"})
		return
	}

	// Insert the note
	query := `
        INSERT INTO notes (id, project_id, timestamp, content, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := db.DB.Exec(
		query,
		uuid.New().String(),
		projectID,
		req.Timestamp,
		req.Content,
		time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add note: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note added successfully"})
}

// ApproveProject marks the project as approved
func ApproveProject(c *gin.Context) {
	projectID := c.Param("id")

	// Check user role
	userRoleInterface, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}
	userRole := userRoleInterface.(string)

	if userRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can approve projects"})
		return
	}

	// Update project status
	_, err := db.DB.Exec(`
        UPDATE projects
        SET status = 'approved', updated_at = now()
        WHERE id = $1
    `, projectID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve project: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project approved successfully"})
}

// RejectProject marks the project as rejected
func RejectProject(c *gin.Context) {
	projectID := c.Param("id")

	// Check user role
	userRoleInterface, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}
	userRole := userRoleInterface.(string)

	if userRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can reject projects"})
		return
	}

	// Update project status
	_, err := db.DB.Exec(`
        UPDATE projects
        SET status = 'rejected', updated_at = now()
        WHERE id = $1
    `, projectID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject project: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project rejected successfully"})
}
