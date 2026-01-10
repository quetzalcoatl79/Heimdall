package handlers

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/nxo/engine/internal/services"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	service *services.AuthService
	render  *render.Engine
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(service *services.AuthService, r *render.Engine) *AuthHandler {
	return &AuthHandler{service: service, render: r}
}

// Register handles user registration
func (h *AuthHandler) Register(c buffalo.Context) error {
	var input services.RegisterInput
	if err := c.Bind(&input); err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid request body"}))
	}

	if input.Email == "" || input.Password == "" {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Email and password are required"}))
	}

	user, err := h.service.Register(c.Request().Context(), input)
	if err != nil {
		if err == services.ErrUserExists {
			return c.Render(409, h.render.JSON(map[string]string{"error": "User already exists"}))
		}
		return c.Render(500, h.render.JSON(map[string]string{"error": "Internal server error"}))
	}

	return c.Render(201, h.render.JSON(map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
	}))
}

// Login handles user login
func (h *AuthHandler) Login(c buffalo.Context) error {
	var input services.LoginInput
	if err := c.Bind(&input); err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid request body"}))
	}

	tokens, user, err := h.service.Login(c.Request().Context(), input)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return c.Render(401, h.render.JSON(map[string]string{"error": "Invalid credentials"}))
		}
		return c.Render(500, h.render.JSON(map[string]string{"error": "Internal server error"}))
	}

	return c.Render(200, h.render.JSON(map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		},
	}))
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c buffalo.Context) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&input); err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid request body"}))
	}

	tokens, err := h.service.RefreshTokens(c.Request().Context(), input.RefreshToken)
	if err != nil {
		return c.Render(401, h.render.JSON(map[string]string{"error": "Invalid refresh token"}))
	}

	return c.Render(200, h.render.JSON(tokens))
}

// Logout handles user logout
func (h *AuthHandler) Logout(c buffalo.Context) error {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.Bind(&input)

	// Get token ID from context if available
	tokenID, _ := c.Value("token_id").(string)

	h.service.Logout(c.Request().Context(), input.RefreshToken, tokenID)

	return c.Render(200, h.render.JSON(map[string]string{"message": "Logged out successfully"}))
}
