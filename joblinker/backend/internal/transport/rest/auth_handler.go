package rest

import (
	"log"
	"net/http"
	"time"

	"joblinker/internal/middleware"
	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication: register, login, refresh.
// Uses the same repository/service layer as the old handler for now,
// with a cleaner package structure that can be swapped later.
type AuthHandler struct {
	userRepo    *repository.UserRepository
	securitySvc *service.SecurityService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(userRepo *repository.UserRepository, securitySvc *service.SecurityService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, securitySvc: securitySvc}
}

// RegisterRequest is the JSON payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=seeker recruiter"`
}

// UserResponse is the JSON response for user data.
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at,omitempty"`
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.userRepo.GetByEmail(req.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}
	user := &model.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         model.UserRole(req.Role),
		TenantID:     uuid.New(),
	}
	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	token := h.generateToken(user)
	// Log registration for security audit
	go func() {
		details := map[string]interface{}{"role": req.Role}
		if err := h.securitySvc.LogEvent(user.ID, "user_registered", details, c.ClientIP()); err != nil {
			log.Printf("Failed to log security event: %v", err)
		}
	}()
	c.JSON(http.StatusCreated, gin.H{
		"user": UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Role:      string(user.Role),
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"token": token,
	})
}

// LoginRequest is the JSON payload for login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		// Log failed login attempt
		go func() {
			details := map[string]interface{}{"email": req.Email}
			_ = h.securitySvc.LogEvent(uuid.Nil, "login_failed", details, c.ClientIP())
		}()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	token := h.generateToken(user)
	c.JSON(http.StatusOK, gin.H{
		"user": UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Role:      string(user.Role),
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"token": token,
	})
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	role := model.UserRole(c.GetString("role"))
	token := h.generateToken(&model.User{
		ID:   userID,
		Role: role,
	})
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) generateToken(user *model.User) string {
	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"role":      string(user.Role),
		"tenant_id": user.TenantID.String(),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middleware.GetJwtSecret())
	if err != nil {
		log.Printf("Failed to sign token: %v", err)
		return ""
	}
	return tokenString
}
