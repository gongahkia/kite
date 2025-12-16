package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gongahkia/kite/internal/api/middleware"
	"github.com/gongahkia/kite/internal/observability"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	logger     *observability.Logger
	authConfig *middleware.AuthConfig
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(logger *observability.Logger, authConfig *middleware.AuthConfig) *AuthHandler {
	return &AuthHandler{
		logger:     logger,
		authConfig: authConfig,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	ClientID  string    `json:"client_id"`
	Roles     []string  `json:"roles"`
}

// Login handles login requests and issues JWT tokens
//
// @Summary User login
// @Description Authenticate with username/password and receive a JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Failed to parse login request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// TODO: Validate credentials against database
	// This is a placeholder implementation
	// In production, you would:
	// 1. Hash the password
	// 2. Query the database for the user
	// 3. Compare password hashes
	// 4. Check if user is active/verified

	// For demo purposes, accept any username/password and create a token
	userID := req.Username
	clientID := "client_" + req.Username
	roles := []string{"user", "read"}

	// Check if username is admin
	if req.Username == "admin" {
		roles = append(roles, "admin", "write")
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(userID, clientID, roles, h.authConfig)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to generate JWT token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	expiresAt := time.Now().Add(h.authConfig.JWTExpiration)

	h.logger.WithFields(map[string]interface{}{
		"user_id":   userID,
		"client_id": clientID,
	}).Info("User logged in successfully")

	return c.JSON(LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
		ClientID:  clientID,
		Roles:     roles,
	})
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	Token string `json:"token" validate:"required"`
}

// RefreshToken handles token refresh requests
//
// @Summary Refresh JWT token
// @Description Refresh an existing JWT token to extend session
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body RefreshRequest true "Current token"
// @Success 200 {object} LoginResponse "Token refreshed"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid or expired token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get user info from context (set by JWT middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	clientID, _ := c.Locals("client_id").(string)
	roles, _ := c.Locals("roles").([]string)

	// Generate new token
	token, err := middleware.GenerateJWT(userID, clientID, roles, h.authConfig)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to refresh JWT token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to refresh token",
		})
	}

	expiresAt := time.Now().Add(h.authConfig.JWTExpiration)

	h.logger.WithFields(map[string]interface{}{
		"user_id":   userID,
		"client_id": clientID,
	}).Info("Token refreshed successfully")

	return c.JSON(LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
		ClientID:  clientID,
		Roles:     roles,
	})
}

// ValidateToken validates a JWT token
//
// @Summary Validate JWT token
// @Description Check if a JWT token is valid
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Token is valid"
// @Failure 401 {object} map[string]interface{} "Invalid or expired token"
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	clientID, _ := c.Locals("client_id").(string)
	roles, _ := c.Locals("roles").([]string)
	authMethod, _ := c.Locals("auth_method").(string)

	return c.JSON(fiber.Map{
		"valid":       true,
		"user_id":     userID,
		"client_id":   clientID,
		"roles":       roles,
		"auth_method": authMethod,
	})
}

// GenerateAPIKeyRequest represents an API key generation request
type GenerateAPIKeyRequest struct {
	ClientID    string `json:"client_id" validate:"required"`
	Description string `json:"description"`
}

// GenerateAPIKeyResponse represents an API key generation response
type GenerateAPIKeyResponse struct {
	APIKey    string    `json:"api_key"`
	ClientID  string    `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
}

// GenerateAPIKey generates a new API key
//
// @Summary Generate API key
// @Description Generate a new API key for a client
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body GenerateAPIKeyRequest true "Client information"
// @Success 200 {object} GenerateAPIKeyResponse "API key generated"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/api-key [post]
func (h *AuthHandler) GenerateAPIKey(c *fiber.Ctx) error {
	// Check if user has admin role
	roles, _ := c.Locals("roles").([]string)
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin role required to generate API keys",
		})
	}

	var req GenerateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate API key (in production, use crypto/rand for secure random generation)
	apiKey := generateRandomString(32)

	// Store in config (in production, store in database)
	h.authConfig.APIKeys[apiKey] = req.ClientID

	h.logger.WithFields(map[string]interface{}{
		"client_id": req.ClientID,
	}).Info("API key generated")

	return c.JSON(GenerateAPIKeyResponse{
		APIKey:    apiKey,
		ClientID:  req.ClientID,
		CreatedAt: time.Now(),
	})
}

// generateRandomString generates a random string (placeholder - use crypto/rand in production)
func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond) // Ensure unique values
	}
	return string(result)
}
