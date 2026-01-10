package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nxo/engine/internal/cache"
	"github.com/nxo/engine/internal/config"
	"github.com/nxo/engine/internal/database"
	"github.com/nxo/engine/internal/middleware"
	"github.com/nxo/engine/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService handles authentication
type AuthService struct {
	db     *database.DB
	cache  *cache.Redis
	config *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(db *database.DB, cache *cache.Redis, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cache: cache, config: cfg}
}

// RegisterInput for user registration
type RegisterInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginInput for user login
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*models.User, error) {
	// Check if user exists
	var existing models.User
	if err := s.db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return nil, ErrUserExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:     input.Email,
		Password:  string(hash),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Role:      "user",
		IsActive:  true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, *models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Update last login
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", now)

	// Generate tokens
	tokens, err := s.generateTokens(&user)
	if err != nil {
		return nil, nil, err
	}

	return tokens, &user, nil
}

// RefreshTokens generates new tokens from a refresh token
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Verify refresh token exists and is valid
	var token models.RefreshToken
	if err := s.db.Where("token = ? AND revoked_at IS NULL AND expires_at > ?", refreshToken, time.Now()).First(&token).Error; err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	var user models.User
	if err := s.db.First(&user, token.UserID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// Revoke old token
	s.db.Model(&token).Update("revoked_at", time.Now())

	// Generate new tokens
	return s.generateTokens(&user)
}

// Logout revokes the refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string, accessTokenID string) error {
	// Revoke refresh token
	s.db.Model(&models.RefreshToken{}).Where("token = ?", refreshToken).Update("revoked_at", time.Now())

	// Blacklist access token
	if accessTokenID != "" {
		s.cache.BlacklistToken(ctx, accessTokenID, s.config.JWT.Expiry)
	}

	return nil
}

func (s *AuthService) generateTokens(user *models.User) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.config.JWT.Expiry)
	tokenID := uuid.New().String()

	// Access token
	claims := middleware.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			Issuer:    "engine",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshTokenString := uuid.New().String()
	refreshExpiry := now.Add(s.config.JWT.RefreshExpiry)

	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: refreshExpiry,
	}
	if err := s.db.Create(refreshTokenModel).Error; err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
	}, nil
}
