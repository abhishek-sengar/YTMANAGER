package api

import (
	"net/http"
	"os"
	"time"

	"github.com/abhishek-sengar/ytmanager/internal/db"
	_ "github.com/abhishek-sengar/ytmanager/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// SignupRequest defines the expected request payload
type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"` // "editor" or "owner"
}

// Signup handler
func Signup(c *gin.Context) {
	var req SignupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert user into database
	query := `
		INSERT INTO users (id, name, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = db.DB.Exec(
		query,
		uuid.New().String(),
		req.Name,
		req.Email,
		string(hashedPassword),
		req.Role,
		time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful"})
}

// LoginRequest is the expected payload for /login
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is returned on successful auth
type LoginResponse struct {
	Token string `json:"token"`
}

// Login handler
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Lookup user
	var (
		id           string
		passwordHash string
		role         string
		name         string
	)
	err := db.DB.QueryRow(
		`SELECT id, password_hash, role, name FROM users WHERE email = $1`,
		req.Email,
	).Scan(&id, &passwordHash, &role, &name)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email "})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// Create JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
		return
	}

	claims := jwt.MapClaims{
		"sub":   id,
		"email": req.Email,
		"name":  name,
		"role":  role,
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
}
