package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/abhishek-sengar/ytmanager/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
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
// func GetProjects(c *gin.Context) {
// 	userIDInterface, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
// 		return
// 	}
// 	userID := userIDInterface.(string)

// 	// Fetch projects where the user is Editor or Owner
// 	rows, err := db.DB.Query(`
//         SELECT id, title, description, video_path, status, created_at, updated_at
//         FROM projects
//         WHERE editor_id = $1 OR owner_id = $1
//         ORDER BY created_at DESC
//     `, userID)

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects: " + err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	var projects []Project
// 	for rows.Next() {
// 		var p Project
// 		if err := rows.Scan(
// 			&p.ID, &p.Title, &p.Description, &p.VideoPath, &p.Status,
// 			&p.CreatedAt, &p.UpdatedAt,
// 		); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse project: " + err.Error()})
// 			return
// 		}
// 		projects = append(projects, p)
// 	}

// 	c.JSON(http.StatusOK, projects)
// }

func GetUserProjects(c *gin.Context) {
	userID, userOk := c.Get("userID")
	role, roleOk := c.Get("userRole")

	if !userOk || !roleOk {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var query string
	switch role {
	case "editor":
		query = `
			SELECT id, title, description, video_path, status, created_at, updated_at
			FROM projects
			WHERE editor_id = $1
			ORDER BY created_at DESC`
	case "owner":
		query = `
			SELECT id, title, description, video_path, status, created_at, updated_at
			FROM projects
			WHERE owner_id = $1
			ORDER BY created_at DESC`
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Unsupported role"})
		return
	}

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed: " + err.Error()})
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.VideoPath, &p.Status,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan failed: " + err.Error()})
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

func GetProjectDetailsByID(c *gin.Context) {
	// Fetch userID from context (JWT middleware should set this)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}
	userID := userIDInterface.(string)

	// Get the project ID from the URL
	projectID := c.Param("id")

	// Declare a variable to hold the project details
	var project Project

	// Query the database using pgx (PostgreSQL)
	// QueryRow will return a single row based on the project ID
	err := db.DB.QueryRow(
		`SELECT id, title, description, video_path, status, editor_id, owner_id, created_at, updated_at
		 FROM projects
		 WHERE id = $1 AND (editor_id = $2 OR owner_id = $2)`,
		projectID, userID).Scan(
		&project.ID, &project.Title, &project.Description, &project.VideoPath,
		&project.Status, &project.EditorID, &project.OwnerID, &project.CreatedAt, &project.UpdatedAt,
	)

	// If there is no project, return 404
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project: " + err.Error()})
		}
		return
	}

	// Return the project details in JSON
	c.JSON(http.StatusOK, project)
}

func GetRecentProjects(c *gin.Context) {
	userID, ok := c.Get("userID")
	role, rok := c.Get("userRole")
	if !ok || !rok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ownerID := c.Query("owner_id")
	channelID := c.Query("channel_id")

	args := []interface{}{userID}
	argIndex := 2
	var query string

	// Determine query based on user role
	if role == "editor" {
		query = `
			SELECT p.id, p.title, p.description, p.status,
				   ch.name AS channel_name,
				   u.name AS owner_name,
				   p.created_at, p.updated_at
			FROM projects p
			JOIN channels ch ON p.channel_id = ch.id
			JOIN users u ON p.owner_id = u.id
			WHERE p.editor_id = $1`
	} else if role == "owner" {
		query = `t4a
			SELECT p.id, p.title, p.description, p.status,
				   ch.name AS channel_name,
				   u.name AS editor_name,
				   p.created_at, p.updated_at
			FROM projects p
			JOIN channels ch ON p.channel_id = ch.id
			JOIN users u ON p.editor_id = u.id
			WHERE p.owner_id = $1`
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unsupported role"})
		return
	}

	// Append extra filters if provided (owner_id or channel_id)
	if channelID != "" {
		query += fmt.Sprintf(" AND p.channel_id = $%d", argIndex)
		args = append(args, channelID)
		argIndex++
	} else if ownerID != "" {
		query += fmt.Sprintf(" AND p.owner_id = $%d", argIndex)
		args = append(args, ownerID)
		argIndex++
	}

	query += " ORDER BY p.updated_at DESC LIMIT 20" // Limit added to avoid huge result sets

	// Run query
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed: " + err.Error()})
		return
	}
	defer rows.Close()

	// Prepare response
	var projects []ProjectResponse
	for rows.Next() {
		var p ProjectResponse
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Status,
			&p.ChannelName, &p.OwnerName, // or editorName, based on role
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan failed: " + err.Error()})
			return
		}
		projects = append(projects, p)
	}

	c.JSON(http.StatusOK, projects)
}

// VideoUploadRequest represents the request body for video upload
type VideoUploadRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	ChannelID   string `json:"channel_id" binding:"required"`
	Privacy     string `json:"privacy" binding:"required"` // "private", "unlisted", "public"
}

// VideoUploadResponse represents the response for video upload
type VideoUploadResponse struct {
	VideoID   string `json:"video_id"`
	UploadURL string `json:"upload_url"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// Initialize GCS client
func initGCSClient() (*storage.Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %v", err)
	}
	return client, nil
}

// UploadVideoToGCS handles the video upload to Google Cloud Storage
// func UploadVideoToGCS(c *gin.Context) {
// 	// Get user ID from context
// 	userID := c.GetString("userID")
// 	if userID == "" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}

// 	// Parse request body
// 	var req VideoUploadRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get the uploaded file
// 	file, header, err := c.Request.FormFile("video")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
// 		return
// 	}
// 	defer file.Close()

// 	// Initialize GCS client
// 	gcsClient, err := initGCSClient()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer gcsClient.Close()

// 	// Create a unique filename
// 	filename := fmt.Sprintf("%s_%d%s", userID, time.Now().UnixNano(), filepath.Ext(header.Filename))
// 	bucketName := os.Getenv("GCS_BUCKET_NAME")
// 	if bucketName == "" {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "GCS bucket not configured"})
// 		return
// 	}

// 	// Create GCS object
// 	obj := gcsClient.Bucket(bucketName).Object(filename)
// 	w := obj.NewWriter(context.Background())

// 	// Copy file to GCS
// 	if _, err := io.Copy(w, file); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload to GCS: %v", err)})
// 		return
// 	}

// 	// Close the writer
// 	if err := w.Close(); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to close GCS writer: %v", err)})
// 		return
// 	}

// 	// Generate signed URL for YouTube upload
// 	opts := &storage.SignedURLOptions{
// 		Scheme:  storage.SigningSchemeV4,
// 		Method:  "GET",
// 		Expires: time.Now().Add(24 * time.Hour),
// 	}

// 	url, err := storage.SignedURL(bucketName, filename, opts)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate signed URL: %v", err)})
// 		return
// 	}

//		// Return response with upload URL
//		c.JSON(http.StatusOK, VideoUploadResponse{
//			UploadURL: url,
//			Status:    "pending",
//			Message:   "Video uploaded to GCS successfully",
//		})
//	}
func UploadVideoToGCS(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get the uploaded video file
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}
	defer file.Close()

	// Initialize GCS client
	gcsClient, err := initGCSClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer gcsClient.Close()

	// Generate filename
	filename := fmt.Sprintf("%s_%d%s", userID, time.Now().UnixNano(), filepath.Ext(header.Filename))
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GCS bucket not configured"})
		return
	}

	// Upload to GCS
	obj := gcsClient.Bucket(bucketName).Object(filename)
	w := obj.NewWriter(context.Background())
	if _, err := io.Copy(w, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload to GCS: %v", err)})
		return
	}
	if err := w.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to close GCS writer: %v", err)})
		return
	}

	// Load service account JSON key
	keyPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read service account key"})
		return
	}

	// Extract credentials
	var creds struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}
	if err := json.Unmarshal(keyData, &creds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse service account key"})
		return
	}

	// Generate signed URL
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(24 * time.Hour),
		GoogleAccessID: creds.ClientEmail,
		PrivateKey:     []byte(creds.PrivateKey),
	}
	url, err := storage.SignedURL(bucketName, filename, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate signed URL: %v", err)})
		return
	}

	// Send response
	c.JSON(http.StatusOK, VideoUploadResponse{
		VideoID:   filename,
		UploadURL: url,
		Status:    "pending",
		Message:   "Video uploaded to GCS successfully",
	})
}

// UploadVideoToYouTube handles the final upload to YouTube
func UploadVideoToYouTube(c *gin.Context) {
	// Get user ID from context
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var req VideoUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get YouTube service
	ytService, err := getYouTubeService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create YouTube video resource
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       req.Title,
			Description: req.Description,
			CategoryId:  "22", // People & Blogs
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: req.Privacy,
		},
	}

	// Call YouTube API to upload
	call := ytService.Videos.Insert([]string{"snippet", "status"}, video)
	_, err = call.Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload to YouTube: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video uploaded to YouTube successfully"})
}

// getYouTubeService creates a YouTube service client
func getYouTubeService(c *gin.Context) (*youtube.Service, error) {
	userID := c.GetString("userID")
	if userID == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	// Get access token from database
	var accessToken string
	err := db.DB.QueryRow(`
		SELECT access_token 
		FROM youtube_accounts 
		WHERE user_id = $1 
		ORDER BY updated_at DESC 
		LIMIT 1
	`, userID).Scan(&accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	// Create OAuth2 config
	cfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{youtube.YoutubeUploadScope},
		Endpoint:     google.Endpoint,
	}

	// Create token source
	token := &oauth2.Token{AccessToken: accessToken}
	client := cfg.Client(context.Background(), token)

	// Create YouTube service
	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %v", err)
	}

	return service, nil
}
