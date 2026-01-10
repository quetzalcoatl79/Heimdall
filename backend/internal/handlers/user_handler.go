package handlers

import (
	"strconv"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/google/uuid"
	"github.com/nxo/engine/internal/services"
)

// UserHandler handles user endpoints
type UserHandler struct {
	service *services.UserService
	render  *render.Engine
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *services.UserService, r *render.Engine) *UserHandler {
	return &UserHandler{service: service, render: r}
}

// Me returns the current user
func (h *UserHandler) Me(c buffalo.Context) error {
	userID, _ := c.Value("user_id").(string)
	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Render(401, h.render.JSON(map[string]string{"error": "Unauthorized"}))
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		return c.Render(404, h.render.JSON(map[string]string{"error": "User not found"}))
	}

	return c.Render(200, h.render.JSON(user))
}

// UpdateMe updates the current user
func (h *UserHandler) UpdateMe(c buffalo.Context) error {
	userID, _ := c.Value("user_id").(string)
	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Render(401, h.render.JSON(map[string]string{"error": "Unauthorized"}))
	}

	var input services.UpdateUserInput
	if err := c.Bind(&input); err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid request body"}))
	}

	// Users can't change their own role or active status
	input.Role = nil
	input.IsActive = nil

	user, err := h.service.Update(id, input)
	if err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to update user"}))
	}

	return c.Render(200, h.render.JSON(user))
}

// List returns a paginated list of users (admin only)
func (h *UserHandler) List(c buffalo.Context) error {
	page, _ := strconv.Atoi(c.Param("page"))
	pageSize, _ := strconv.Atoi(c.Param("page_size"))
	role := c.Param("role")
	search := c.Param("search")

	result, err := h.service.List(services.UserListParams{
		Page:     page,
		PageSize: pageSize,
		Role:     role,
		Search:   search,
	})
	if err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to list users"}))
	}

	return c.Render(200, h.render.JSON(result))
}

// Get returns a user by ID (admin only)
func (h *UserHandler) Get(c buffalo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid user ID"}))
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		return c.Render(404, h.render.JSON(map[string]string{"error": "User not found"}))
	}

	return c.Render(200, h.render.JSON(user))
}

// Create creates a new user (admin only)
func (h *UserHandler) Create(c buffalo.Context) error {
	var input services.CreateUserInput
	if err := c.Bind(&input); err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid request body"}))
	}

	user, err := h.service.Create(input)
	if err != nil {
		if err == services.ErrUserExists {
			return c.Render(409, h.render.JSON(map[string]string{"error": "User already exists"}))
		}
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to create user"}))
	}

	return c.Render(201, h.render.JSON(user))
}

// Update updates a user (admin only)
func (h *UserHandler) Update(c buffalo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid user ID"}))
	}

	var input services.UpdateUserInput
	if err := c.Bind(&input); err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid request body"}))
	}

	user, err := h.service.Update(id, input)
	if err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to update user"}))
	}

	return c.Render(200, h.render.JSON(user))
}

// Delete deletes a user (admin only)
func (h *UserHandler) Delete(c buffalo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.Render(400, h.render.JSON(map[string]string{"error": "Invalid user ID"}))
	}

	if err := h.service.Delete(id); err != nil {
		return c.Render(500, h.render.JSON(map[string]string{"error": "Failed to delete user"}))
	}

	return c.Render(200, h.render.JSON(map[string]string{"message": "User deleted successfully"}))
}
