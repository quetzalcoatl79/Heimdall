package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User represents an application user
type User struct {
	BaseModel
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"-" json:"-"`
	Password     string     `gorm:"column:password_hash;not null" json:"-"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Role         string     `gorm:"default:user" json:"role"` // admin, worker, user
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	Metadata     JSON       `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`
}

// TableName overrides table name
func (User) TableName() string {
	return "users"
}

// Role represents user roles for RBAC
type Role struct {
	BaseModel
	Name        string   `gorm:"uniqueIndex;not null" json:"name"`
	Description string   `json:"description"`
	Permissions []string `gorm:"type:text[];default:'{}'" json:"permissions"`
}

func (Role) TableName() string {
	return "roles"
}

// Plugin represents an installed plugin
type Plugin struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;not null" json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Enabled     bool   `gorm:"default:false" json:"enabled"`
	Config      JSON   `gorm:"type:jsonb;default:'{}'" json:"config"`
	Manifest    JSON   `gorm:"type:jsonb;default:'{}'" json:"manifest"`
	InstalledAt time.Time `json:"installed_at"`
}

func (Plugin) TableName() string {
	return "plugins"
}

// Job represents a background job
type Job struct {
	BaseModel
	Queue     string     `gorm:"index;not null" json:"queue"`
	Type      string     `gorm:"not null" json:"type"`
	Payload   JSON       `gorm:"type:jsonb;default:'{}'" json:"payload"`
	Status    string     `gorm:"default:pending;index" json:"status"` // pending, running, completed, failed
	Attempts  int        `gorm:"default:0" json:"attempts"`
	MaxRetries int       `gorm:"default:3" json:"max_retries"`
	Error     string     `json:"error,omitempty"`
	RunAt     *time.Time `json:"run_at,omitempty"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (Job) TableName() string {
	return "jobs"
}

// AuditLog tracks important actions
type AuditLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action    string    `gorm:"not null" json:"action"`
	Resource  string    `json:"resource"`
	ResourceID string   `json:"resource_id,omitempty"`
	Details   JSON      `gorm:"type:jsonb;default:'{}'" json:"details"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// RefreshToken for JWT refresh mechanism
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
