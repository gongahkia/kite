package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/pkg/errors"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	APIKeys       map[string]string // API Key -> User/Client ID
	JWTSecret     string
	JWTExpiration time.Duration
	Skipper       func(*fiber.Ctx) bool // Optional skip function
}

// DefaultAuthConfig returns a default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		APIKeys:       make(map[string]string),
		JWTExpiration: 24 * time.Hour,
		Skipper:       nil,
	}
}

// APIKeyAuth validates API key from headers
func APIKeyAuth(config *AuthConfig, logger *observability.Logger) fiber.Handler {
	if config == nil {
		config = DefaultAuthConfig()
	}

	return func(c *fiber.Ctx) error {
		// Skip if skipper function returns true
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Get API key from header
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			// Also try Authorization header with Bearer scheme
			auth := c.Get("Authorization")
			if strings.HasPrefix(auth, "ApiKey ") {
				apiKey = strings.TrimPrefix(auth, "ApiKey ")
			}
		}

		if apiKey == "" {
			logger.WithField("path", c.Path()).Warn("Missing API key")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing API key",
				"hint":  "Provide X-API-Key header or Authorization: ApiKey <key>",
			})
		}

		// Validate API key
		clientID, valid := config.APIKeys[apiKey]
		if !valid {
			logger.WithFields(map[string]interface{}{
				"path":    c.Path(),
				"api_key": maskAPIKey(apiKey),
			}).Warn("Invalid API key")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
			})
		}

		// Store client ID in context for later use
		c.Locals("client_id", clientID)
		c.Locals("auth_method", "api_key")

		logger.WithFields(map[string]interface{}{
			"client_id": clientID,
			"path":      c.Path(),
		}).Debug("API key authentication successful")

		return c.Next()
	}
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	ClientID string   `json:"client_id"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTAuth validates JWT tokens
func JWTAuth(config *AuthConfig, logger *observability.Logger) fiber.Handler {
	if config == nil {
		config = DefaultAuthConfig()
	}

	if config.JWTSecret == "" {
		logger.Fatal("JWT secret not configured")
	}

	return func(c *fiber.Ctx) error {
		// Skip if skipper function returns true
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Get token from Authorization header
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			logger.WithField("path", c.Path()).Warn("Missing Bearer token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid authorization header",
				"hint":  "Provide Authorization: Bearer <token>",
			})
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, &errors.KiteError{
					Code:    "INVALID_TOKEN",
					Message: "Invalid signing method",
				}
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil {
			logger.WithFields(map[string]interface{}{
				"path":  c.Path(),
				"error": err.Error(),
			}).Warn("JWT validation failed")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		if !token.Valid {
			logger.WithField("path", c.Path()).Warn("Invalid JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			logger.WithField("path", c.Path()).Error("Failed to parse JWT claims")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("client_id", claims.ClientID)
		c.Locals("roles", claims.Roles)
		c.Locals("auth_method", "jwt")

		logger.WithFields(map[string]interface{}{
			"user_id":   claims.UserID,
			"client_id": claims.ClientID,
			"path":      c.Path(),
		}).Debug("JWT authentication successful")

		return c.Next()
	}
}

// OptionalAuth tries JWT first, then API key, allows unauthenticated if both fail
func OptionalAuth(config *AuthConfig, logger *observability.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try JWT first
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenString := strings.TrimPrefix(auth, "Bearer ")
			token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, &errors.KiteError{Code: "INVALID_TOKEN", Message: "Invalid signing method"}
				}
				return []byte(config.JWTSecret), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*JWTClaims); ok {
					c.Locals("user_id", claims.UserID)
					c.Locals("client_id", claims.ClientID)
					c.Locals("roles", claims.Roles)
					c.Locals("auth_method", "jwt")
					c.Locals("authenticated", true)
					return c.Next()
				}
			}
		}

		// Try API key
		apiKey := c.Get("X-API-Key")
		if apiKey == "" && strings.HasPrefix(auth, "ApiKey ") {
			apiKey = strings.TrimPrefix(auth, "ApiKey ")
		}

		if apiKey != "" {
			if clientID, valid := config.APIKeys[apiKey]; valid {
				c.Locals("client_id", clientID)
				c.Locals("auth_method", "api_key")
				c.Locals("authenticated", true)
				return c.Next()
			}
		}

		// Allow unauthenticated
		c.Locals("authenticated", false)
		return c.Next()
	}
}

// RequireRoles checks if user has required roles (only works with JWT)
func RequireRoles(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roles, ok := c.Locals("roles").([]string)
		if !ok || roles == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
			})
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, required := range requiredRoles {
			for _, role := range roles {
				if role == required {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":          "Insufficient permissions",
				"required_roles": requiredRoles,
			})
		}

		return c.Next()
	}
}

// GenerateJWT generates a new JWT token
func GenerateJWT(userID, clientID string, roles []string, config *AuthConfig) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		ClientID: clientID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kite-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// maskAPIKey masks an API key for logging (shows first 8 chars)
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "..." + strings.Repeat("*", len(key)-8)
}
