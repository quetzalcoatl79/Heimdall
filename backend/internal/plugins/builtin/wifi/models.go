package wifi

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// WiFi Plugin Models
// ============================================================================
// These models are specific to the WiFi plugin and are auto-migrated
// when the plugin is loaded.

// WifiCapture represents a WiFi pentest capture session
type WifiCapture struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	SSID            string         `gorm:"not null" json:"ssid"`
	BSSID           string         `gorm:"not null" json:"bssid"`
	Channel         int            `json:"channel"`
	Security        string         `json:"security"`
	CapturePath     string         `gorm:"not null" json:"capture_path"`
	CaptureName     string         `gorm:"not null" json:"capture_name"`
	FileSize        int64          `gorm:"default:0" json:"file_size"`
	HasHandshake    bool           `gorm:"default:false" json:"has_handshake"`
	InterfaceUsed   string         `json:"interface_used"`
	DurationSeconds int            `json:"duration_seconds"`
	StartedAt       *time.Time     `json:"started_at"`
	EndedAt         *time.Time     `json:"ended_at"`
	Status          string         `gorm:"default:running" json:"status"` // running, completed, stopped, failed
	Cracked         bool           `gorm:"default:false" json:"cracked"`
	CrackedPassword string         `json:"cracked_password,omitempty"`
	CrackedAt       *time.Time     `json:"cracked_at,omitempty"`
	Notes           string         `json:"notes,omitempty"`
}

func (WifiCapture) TableName() string {
	return "wifi_captures"
}

// WifiNetwork represents a detected WiFi network (for caching scan results)
type WifiNetwork struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	SSID      string         `gorm:"index" json:"ssid"`
	BSSID     string         `gorm:"uniqueIndex" json:"bssid"`
	Channel   int            `json:"channel"`
	Signal    int            `json:"signal"`
	Security  string         `json:"security"`
	Vendor    string         `json:"vendor,omitempty"`
	ISP       string         `json:"isp,omitempty"`
	WPS       bool           `gorm:"default:false" json:"wps"`
	FirstSeen time.Time      `json:"first_seen"`
	LastSeen  time.Time      `json:"last_seen"`
}

func (WifiNetwork) TableName() string {
	return "wifi_networks"
}

// WifiBruteforceResult represents a bruteforce attack result
type WifiBruteforceResult struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	CaptureID      *uuid.UUID     `gorm:"type:uuid" json:"capture_id,omitempty"`
	Capture        *WifiCapture   `gorm:"foreignKey:CaptureID" json:"capture,omitempty"`
	SSID           string         `json:"ssid"`
	BSSID          string         `json:"bssid"`
	CapturePath    string         `json:"capture_path"`
	WordlistPath   string         `json:"wordlist_path"`
	WordlistName   string         `json:"wordlist_name"`
	Success        bool           `gorm:"default:false" json:"success"`
	Password       string         `json:"password,omitempty"`
	KeysTested     int64          `json:"keys_tested"`
	KeysTotal      int64          `json:"keys_total"`
	DurationSecs   float64        `json:"duration_seconds"`
	StartedAt      time.Time      `json:"started_at"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	Status         string         `gorm:"default:running" json:"status"` // running, completed, stopped, failed
	ErrorMessage   string         `json:"error_message,omitempty"`
}

func (WifiBruteforceResult) TableName() string {
	return "wifi_bruteforce_results"
}

// WifiDeauthLog represents a deauth attack log entry
type WifiDeauthLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	TargetBSSID  string         `gorm:"not null" json:"target_bssid"`
	TargetSSID   string         `json:"target_ssid"`
	ClientMAC    string         `json:"client_mac,omitempty"` // FF:FF:FF:FF:FF:FF for broadcast
	Channel      int            `json:"channel"`
	PacketsSent  int            `json:"packets_sent"`
	Interface    string         `json:"interface"`
	DurationSecs float64        `json:"duration_seconds"`
}

func (WifiDeauthLog) TableName() string {
	return "wifi_deauth_logs"
}

// ============================================================================
// Plugin Model Registration
// ============================================================================

// Models returns all models that need to be auto-migrated for this plugin
func Models() []interface{} {
	return []interface{}{
		&WifiCapture{},
		&WifiNetwork{},
		&WifiBruteforceResult{},
		&WifiDeauthLog{},
	}
}

// AutoMigrate runs auto-migration for all plugin models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(Models()...)
}
