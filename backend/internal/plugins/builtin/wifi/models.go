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
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	CaptureID    *uuid.UUID     `gorm:"type:uuid" json:"capture_id,omitempty"`
	Capture      *WifiCapture   `gorm:"foreignKey:CaptureID" json:"capture,omitempty"`
	SSID         string         `json:"ssid"`
	BSSID        string         `json:"bssid"`
	CapturePath  string         `json:"capture_path"`
	WordlistPath string         `json:"wordlist_path"`
	WordlistName string         `json:"wordlist_name"`
	Success      bool           `gorm:"default:false" json:"success"`
	Password     string         `json:"password,omitempty"`
	KeysTested   int64          `json:"keys_tested"`
	KeysTotal    int64          `json:"keys_total"`
	DurationSecs float64        `json:"duration_seconds"`
	StartedAt    time.Time      `json:"started_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	Status       string         `gorm:"default:running" json:"status"` // running, completed, stopped, failed
	ErrorMessage string         `json:"error_message,omitempty"`
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

// WifiAudit represents a WiFi pentest audit/engagement
type WifiAudit struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	// Client information
	ClientName    string `gorm:"not null" json:"client_name"` // Nom de l'entreprise ou utilisateur
	ClientContact string `json:"client_contact,omitempty"`    // Email/téléphone contact
	Location      string `json:"location,omitempty"`          // Lieu du test
	// Audit metadata
	AuditType  string     `gorm:"default:wifi" json:"audit_type"` // wifi, wpa, wep, wps
	StartDate  time.Time  `json:"start_date"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Status     string     `gorm:"default:in_progress" json:"status"` // in_progress, completed, cancelled
	TesterName string     `json:"tester_name,omitempty"`
	// Results summary
	NetworksScanned    int `gorm:"default:0" json:"networks_scanned"`
	NetworksTested     int `gorm:"default:0" json:"networks_tested"`
	HandshakesCaptured int `gorm:"default:0" json:"handshakes_captured"`
	PasswordsCracked   int `gorm:"default:0" json:"passwords_cracked"`
	// Report
	ReportGenerated   bool       `gorm:"default:false" json:"report_generated"`
	ReportPath        string     `json:"report_path,omitempty"`
	ReportGeneratedAt *time.Time `json:"report_generated_at,omitempty"`
	Notes             string     `gorm:"type:text" json:"notes,omitempty"`
}

func (WifiAudit) TableName() string {
	return "wifi_audits"
}

// WifiAuditNetwork links an audit to tested networks
type WifiAuditNetwork struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	AuditID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"audit_id"`
	Audit     *WifiAudit `gorm:"foreignKey:AuditID" json:"audit,omitempty"`
	SSID      string     `json:"ssid"`
	BSSID     string     `json:"bssid"`
	Channel   int        `json:"channel"`
	Security  string     `json:"security"`
	Vendor    string     `json:"vendor,omitempty"`
	ISP       string     `json:"isp,omitempty"`
	// Test results
	HandshakeCaptured bool       `gorm:"default:false" json:"handshake_captured"`
	CaptureID         *uuid.UUID `gorm:"type:uuid" json:"capture_id,omitempty"`
	PasswordCracked   bool       `gorm:"default:false" json:"password_cracked"`
	Password          string     `json:"password,omitempty"`
	CrackMethod       string     `json:"crack_method,omitempty"` // wordlist, mask, hybrid
	CrackDuration     float64    `json:"crack_duration_seconds,omitempty"`
	BruteforceID      *uuid.UUID `gorm:"type:uuid" json:"bruteforce_id,omitempty"`
	// Vulnerability assessment
	VulnerabilityLevel string `json:"vulnerability_level,omitempty"` // critical, high, medium, low
	Recommendations    string `gorm:"type:text" json:"recommendations,omitempty"`
}

func (WifiAuditNetwork) TableName() string {
	return "wifi_audit_networks"
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
		&WifiAudit{},
		&WifiAuditNetwork{},
	}
}

// AutoMigrate runs auto-migration for all plugin models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(Models()...)
}
