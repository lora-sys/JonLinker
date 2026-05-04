package handler

import (
	"net/http"
	"os"
	"time"

	"joblinker/internal/model"
	"joblinker/internal/repository"
	"joblinker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo    *repository.UserRepository
	securitySvc *service.SecurityService
}

func NewAuthHandler(userRepo *repository.UserRepository, securitySvc *service.SecurityService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, securitySvc: securitySvc}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var jwtSecret = []byte(getEnv("JWT_SECRET", "joblinker-dev-secret-change-in-production"))

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=seeker recruiter"`
}

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
	}
	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	h.securitySvc.LogEvent(user.ID, "register", map[string]interface{}{"email": user.Email}, c.ClientIP())
	token := h.generateToken(user)
	c.JSON(http.StatusCreated, gin.H{"user": user, "token": token})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	h.securitySvc.LogEvent(user.ID, "login", nil, c.ClientIP())
	token := h.generateToken(user)
	c.JSON(http.StatusOK, gin.H{"user": user, "token": token})
}

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
		"sub":  user.ID.String(),
		"role": string(user.Role),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}
