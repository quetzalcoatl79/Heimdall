package captiveportal
package captiveportal

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// BruteforceConfig holds configuration for bruteforce attack
type BruteforceConfig struct {
	TargetURL    string   `json:"target_url"`
	Method       string   `json:"method"`
	Headers      map[string]string `json:"headers,omitempty"`
	FieldName    string   `json:"field_name"`    // code, login, password
	OtherFields  map[string]string `json:"other_fields,omitempty"` // Fixed values for other fields
	
	// Charset configuration
	UseDigits      bool `json:"use_digits"`       // 0-9
	UseLowercase   bool `json:"use_lowercase"`    // a-z
	UseUppercase   bool `json:"use_uppercase"`    // A-Z
	UseSpecial     bool `json:"use_special"`      // !@#$%^&*
	CustomCharset  string `json:"custom_charset,omitempty"` // Custom chars
	
	// Length configuration
	MinLength int `json:"min_length"`
	MaxLength int `json:"max_length"`
	
	// Rate limiting
	RateLimit int `json:"rate_limit"` // requests per second
	MaxAttempts int `json:"max_attempts,omitempty"` // 0 = unlimited
	
	// Success detection
	SuccessPattern   string `json:"success_pattern,omitempty"`   // Regex in response body
	FailurePattern   string `json:"failure_pattern,omitempty"`   // Regex in response body
	SuccessStatus    int    `json:"success_status,omitempty"`    // HTTP status code
	SuccessRedirect  bool   `json:"success_redirect,omitempty"`  // Success if redirect
}

// BruteforceProgress tracks the progress of a bruteforce attack
type BruteforceProgress struct {
	mu            sync.Mutex
	Running       bool      `json:"running"`
	StartedAt     time.Time `json:"started_at"`
	TotalPossible int64     `json:"total_possible"`
	Attempted     int64     `json:"attempted"`
	Successful    int       `json:"successful"`
	Failed        int       `json:"failed"`
	CurrentValue  string    `json:"current_value"`
	FoundCodes    []string  `json:"found_codes,omitempty"`
	LastError     string    `json:"last_error,omitempty"`
	ETA           string    `json:"eta,omitempty"`
	Speed         float64   `json:"speed"` // attempts per second
}

var (
	bruteProgress   BruteforceProgress
	bruteCancel     chan struct{}
	bruteCancelOnce sync.Once
)

// GetCharset builds the charset based on configuration
func GetCharset(config BruteforceConfig) string {
	var charset strings.Builder
	
	if config.UseDigits {
		charset.WriteString("0123456789")
	}
	if config.UseLowercase {
		charset.WriteString("abcdefghijklmnopqrstuvwxyz")
	}
	if config.UseUppercase {
		charset.WriteString("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	if config.UseSpecial {
		charset.WriteString("!@#$%^&*()-_=+[]{}|;:,.<>?")
	}
	if config.CustomCharset != "" {
		charset.WriteString(config.CustomCharset)
	}
	
	// Default to digits if nothing selected
	if charset.Len() == 0 {
		return "0123456789"
	}
	
	// Remove duplicates
	seen := make(map[rune]bool)
	var unique strings.Builder
	for _, r := range charset.String() {
		if !seen[r] {
			seen[r] = true
			unique.WriteRune(r)
		}
	}
	
	return unique.String()
}

// CalculateTotalCombinations returns the total number of possible combinations
func CalculateTotalCombinations(config BruteforceConfig) int64 {
	charset := GetCharset(config)
	charsetLen := int64(len(charset))
	
	var total int64 = 0
	for length := config.MinLength; length <= config.MaxLength; length++ {
		total += int64(math.Pow(float64(charsetLen), float64(length)))
	}
	
	return total
}

// GenerateCombinations generates all combinations for bruteforce
func GenerateCombinations(config BruteforceConfig, callback func(string) bool) {
	charset := GetCharset(config)
	
	for length := config.MinLength; length <= config.MaxLength; length++ {
		if !generateAtLength(charset, length, "", callback) {
			return
		}
	}
}

// generateAtLength generates all combinations of a specific length
func generateAtLength(charset string, length int, current string, callback func(string) bool) bool {
	if len(current) == length {
		return callback(current)
	}
	
	for _, c := range charset {
		if !generateAtLength(charset, length, current+string(c), callback) {
			return false
		}
	}
	return true
}

// GetBruteforceProgress returns the current progress
func GetBruteforceProgress() BruteforceProgress {
	bruteProgress.mu.Lock()
	defer bruteProgress.mu.Unlock()
	
	progress := bruteProgress
	
	// Calculate ETA
	if progress.Running && progress.Speed > 0 {
		remaining := progress.TotalPossible - progress.Attempted
		seconds := float64(remaining) / progress.Speed
		progress.ETA = formatDuration(time.Duration(seconds) * time.Second)
	}
	
	return progress
}

// StopBruteforce stops the running bruteforce
func StopBruteforce() {
	if bruteCancel != nil {
		bruteCancelOnce.Do(func() {
			close(bruteCancel)
		})
	}
	bruteProgress.mu.Lock()
	bruteProgress.Running = false
	bruteProgress.mu.Unlock()
}

// formatDuration formats a duration in a human readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dj %dh", days, hours)
}

// PresetCharsets contains common charset presets
var PresetCharsets = map[string]BruteforceConfig{
	"digits_only": {
		UseDigits: true,
		MinLength: 4,
		MaxLength: 6,
	},
	"hotel_room": {
		UseDigits: true,
		MinLength: 3,
		MaxLength: 4,
	},
	"alphanumeric_lower": {
		UseDigits:    true,
		UseLowercase: true,
		MinLength:    4,
		MaxLength:    6,
	},
	"alphanumeric_mixed": {
		UseDigits:    true,
		UseLowercase: true,
		UseUppercase: true,
		MinLength:    4,
		MaxLength:    8,
	},
	"full": {
		UseDigits:    true,
		UseLowercase: true,
		UseUppercase: true,
		UseSpecial:   true,
		MinLength:    4,
		MaxLength:    8,
	},
}

// EstimateBruteforceDuration estimates how long a bruteforce will take
func EstimateBruteforceDuration(config BruteforceConfig) map[string]any {
	total := CalculateTotalCombinations(config)
	charset := GetCharset(config)
	rateLimit := config.RateLimit
	if rateLimit <= 0 {
		rateLimit = 2
	}
	
	seconds := float64(total) / float64(rateLimit)
	duration := time.Duration(seconds) * time.Second
	
	return map[string]any{
		"charset":        charset,
		"charset_length": len(charset),
		"min_length":     config.MinLength,
		"max_length":     config.MaxLength,
		"total_combinations": total,
		"rate_limit":     rateLimit,
		"estimated_duration": formatDuration(duration),
		"estimated_seconds":  seconds,
	}
}
