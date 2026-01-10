package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user operations
type UserService struct {
	db *database.DB
}

// NewUserService creates a new user service
func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

// UserListParams for listing users
type UserListParams struct {
	Page     int
	PageSize int
	Role     string
	Search   string
}

// UserListResult for paginated results
type UserListResult struct {
	Users      []models.User `json:"users"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// List returns a paginated list of users
func (s *UserService) List(params UserListParams) (*UserListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	if params.Role != "" {
		query = query.Where("role = ?", params.Role)
	}
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", search, search, search)
	}

	query.Count(&total)

	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return &UserListResult{
		Users:      users,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID returns a user by ID
func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// CreateUserInput for creating a user
type CreateUserInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// Create creates a new user (admin only)
func (s *UserService) Create(input CreateUserInput) (*models.User, error) {
	// Check if user exists
	var existing models.User
	if err := s.db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := input.Role
	if role == "" {
		role = "user"
	}

	user := &models.User{
		Email:     input.Email,
		Password:  string(hash),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Role:      role,
		IsActive:  true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserInput for updating a user
type UpdateUserInput struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Role      *string `json:"role,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// Update updates a user
func (s *UserService) Update(id uuid.UUID, input UpdateUserInput) (*models.User, error) {
	user, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if input.FirstName != nil {
		updates["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		updates["last_name"] = *input.LastName
	}
	if input.Role != nil {
		updates["role"] = *input.Role
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if len(updates) > 0 {
		if err := s.db.Model(user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.GetByID(id)
}

// Delete soft deletes a user
func (s *UserService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
