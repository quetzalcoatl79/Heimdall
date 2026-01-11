package wifi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// VendorInfo contains information about a WiFi device vendor
type VendorInfo struct {
	Manufacturer string `json:"manufacturer"`
	ISP          string `json:"isp,omitempty"`
	RouterModel  string `json:"router_model,omitempty"`
	Country      string `json:"country,omitempty"`
}

// PasswordPattern describes the default password format for a router
type PasswordPattern struct {
	ISP           string   `json:"isp"`
	RouterModels  []string `json:"router_models,omitempty"`
	Description   string   `json:"description"`
	Length        string   `json:"length"`         // e.g., "26", "8-12", "40-50"
	CharacterSet  string   `json:"character_set"`  // e.g., "hex", "alphanumeric", "latin-words"
	Format        string   `json:"format"`         // regex or description
	Example       string   `json:"example"`
	GeneratorType string   `json:"generator_type"` // "wordlist", "mask", "hybrid"
	MaskPattern   string   `json:"mask_pattern,omitempty"` // For hashcat masks
	Notes         string   `json:"notes,omitempty"`
}

// Known OUI prefixes for common router manufacturers
// Format: first 3 bytes of MAC (uppercase, colon-separated) -> VendorInfo
var knownOUI = map[string]VendorInfo{
	// Freebox (Free/Iliad)
	"14:0C:76": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"24:8A:07": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"34:27:92": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"68:A3:78": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"70:FC:8F": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"78:94:B4": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"F4:CA:E5": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"E4:9E:12": {Manufacturer: "Freebox", ISP: "Free", Country: "FR"},
	"20:66:CF": {Manufacturer: "Freebox", ISP: "Free", RouterModel: "Freebox Pop/Delta", Country: "FR"},
	"22:66:CF": {Manufacturer: "Freebox", ISP: "Free", RouterModel: "Freebox Pop/Delta", Country: "FR"},
	
	// Orange Livebox
	"00:1E:74": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	"30:23:03": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	"34:62:88": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	"64:7C:34": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	"80:E8:6F": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	"A4:3E:51": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	"E8:AD:A6": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	"2C:39:96": {Manufacturer: "Sagemcom", ISP: "Orange", RouterModel: "Livebox", Country: "FR"},
	
	// SFR Box
	"00:1A:2B": {Manufacturer: "Technicolor", ISP: "SFR", RouterModel: "SFR Box", Country: "FR"},
	"00:26:91": {Manufacturer: "Sagemcom", ISP: "SFR", RouterModel: "SFR Box", Country: "FR"},
	"28:C6:8E": {Manufacturer: "Netgear", ISP: "SFR", RouterModel: "SFR Box", Country: "FR"},
	"44:CE:7D": {Manufacturer: "Sagemcom", ISP: "SFR", RouterModel: "SFR Box", Country: "FR"},
	"74:9D:DC": {Manufacturer: "Sagemcom", ISP: "SFR", RouterModel: "SFR Box 8", Country: "FR"},
	"A8:4E:3F": {Manufacturer: "Sagemcom", ISP: "SFR", RouterModel: "SFR Box", Country: "FR"},
	
	// Bouygues Bbox
	"00:19:70": {Manufacturer: "Sagemcom", ISP: "Bouygues", RouterModel: "Bbox", Country: "FR"},
	"00:1F:9F": {Manufacturer: "Sagemcom", ISP: "Bouygues", RouterModel: "Bbox", Country: "FR"},
	"3C:81:D8": {Manufacturer: "Sagemcom", ISP: "Bouygues", RouterModel: "Bbox", Country: "FR"},
	"5C:A4:8A": {Manufacturer: "Sagemcom", ISP: "Bouygues", RouterModel: "Bbox", Country: "FR"},
	"DC:0B:1A": {Manufacturer: "Sagemcom", ISP: "Bouygues", RouterModel: "Bbox", Country: "FR"},
	
	// Generic routers
	"00:14:BF": {Manufacturer: "Linksys", Country: "US"},
	"00:18:39": {Manufacturer: "Cisco-Linksys", Country: "US"},
	"00:1C:10": {Manufacturer: "Cisco-Linksys", Country: "US"},
	"00:1E:58": {Manufacturer: "D-Link", Country: "TW"},
	"00:22:B0": {Manufacturer: "D-Link", Country: "TW"},
	"00:24:01": {Manufacturer: "D-Link", Country: "TW"},
	"00:26:5A": {Manufacturer: "D-Link", Country: "TW"},
	"1C:7E:E5": {Manufacturer: "D-Link", Country: "TW"},
	"00:1F:33": {Manufacturer: "Netgear", Country: "US"},
	"00:24:B2": {Manufacturer: "Netgear", Country: "US"},
	"00:26:F2": {Manufacturer: "Netgear", Country: "US"},
	"04:A1:51": {Manufacturer: "Netgear", Country: "US"},
	"20:0C:C8": {Manufacturer: "Netgear", Country: "US"},
	"00:1A:92": {Manufacturer: "TP-Link", Country: "CN"},
	"14:CC:20": {Manufacturer: "TP-Link", Country: "CN"},
	"50:C7:BF": {Manufacturer: "TP-Link", Country: "CN"},
	"64:66:B3": {Manufacturer: "TP-Link", Country: "CN"},
	"90:F6:52": {Manufacturer: "TP-Link", Country: "CN"},
	"00:24:17": {Manufacturer: "Asus", Country: "TW"},
	"08:60:6E": {Manufacturer: "Asus", Country: "TW"},
	"10:C3:7B": {Manufacturer: "Asus", Country: "TW"},
}

// Known password patterns for ISPs and routers
var passwordPatterns = map[string]PasswordPattern{
	"Free": {
		ISP:           "Free",
		RouterModels:  []string{"Freebox Revolution", "Freebox Mini 4K", "Freebox Pop", "Freebox Delta"},
		Description:   "4 mots latins séparés par des tirets, avec suffixes aléatoires (chiffres/symboles)",
		Length:        "35-55",
		CharacterSet:  "latin-words + digits + symbols",
		Format:        `^[a-z]+-[a-z0-9%#\*]+-[a-z0-9%#\*]+-[a-z0-9%#\*]+$`,
		Example:       "persordida-peritorum2%-tenoris6-helluo#*",
		GeneratorType: "wordlist-combo",
		Notes:         "Utilise un dictionnaire de ~2000 mots latins. Suffixes: chiffres 0-9, symboles %, #, *",
	},
	"Orange": {
		ISP:           "Orange",
		RouterModels:  []string{"Livebox 4", "Livebox 5", "Livebox 6"},
		Description:   "26 caractères hexadécimaux (anciennes) ou format mixte (nouvelles)",
		Length:        "26",
		CharacterSet:  "hex (0-9, A-F)",
		Format:        `^[0-9A-F]{26}$`,
		Example:       "A1B2C3D4E5F6G7H8I9J0K1L2M3",
		GeneratorType: "mask",
		MaskPattern:   "?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H?H",
		Notes:         "Les nouvelles Livebox utilisent parfois un format différent",
	},
	"SFR": {
		ISP:           "SFR",
		RouterModels:  []string{"SFR Box 7", "SFR Box 8"},
		Description:   "12 caractères alphanumériques majuscules",
		Length:        "12",
		CharacterSet:  "uppercase alphanumeric",
		Format:        `^[A-Z0-9]{12}$`,
		Example:       "ABC123DEF456",
		GeneratorType: "mask",
		MaskPattern:   "?u?u?u?d?d?d?u?u?u?d?d?d",
		Notes:         "Anciennes box: 8 caractères. Nouvelles: 12 caractères",
	},
	"Bouygues": {
		ISP:           "Bouygues",
		RouterModels:  []string{"Bbox Fit", "Bbox Must", "Bbox Ultym"},
		Description:   "8 à 12 caractères alphanumériques",
		Length:        "8-12",
		CharacterSet:  "alphanumeric",
		Format:        `^[A-Za-z0-9]{8,12}$`,
		Example:       "AbCdEf123456",
		GeneratorType: "mask",
		MaskPattern:   "?a?a?a?a?a?a?a?a",
	},
	"TP-Link": {
		ISP:           "TP-Link",
		RouterModels:  []string{"Archer", "Deco"},
		Description:   "8 chiffres (souvent imprimé sur l'étiquette)",
		Length:        "8",
		CharacterSet:  "digits",
		Format:        `^\d{8}$`,
		Example:       "12345678",
		GeneratorType: "mask",
		MaskPattern:   "?d?d?d?d?d?d?d?d",
	},
	"Netgear": {
		ISP:           "Netgear",
		RouterModels:  []string{"Nighthawk", "Orbi"},
		Description:   "Mot + chiffres (format variable)",
		Length:        "10-14",
		CharacterSet:  "word + digits",
		Format:        `^[a-z]+\d+$`,
		Example:       "adjective12345",
		GeneratorType: "hybrid",
	},
}

// LookupVendor returns vendor information for a given BSSID
func LookupVendor(bssid string) VendorInfo {
	// Normalize BSSID: uppercase, colon-separated
	bssid = strings.ToUpper(strings.ReplaceAll(bssid, "-", ":"))
	
	// Extract OUI (first 3 bytes)
	parts := strings.Split(bssid, ":")
	if len(parts) < 3 {
		return VendorInfo{}
	}
	oui := strings.Join(parts[:3], ":")
	
	// Check local cache first
	if info, ok := knownOUI[oui]; ok {
		return info
	}
	
	// Try online lookup (macvendors.com API)
	info := lookupOnline(oui)
	if info.Manufacturer != "" {
		return info
	}
	
	return VendorInfo{Manufacturer: "Unknown"}
}

// lookupOnline queries macvendors.com API for OUI info
func lookupOnline(oui string) VendorInfo {
	// Clean OUI for API (remove colons)
	cleanOUI := strings.ReplaceAll(oui, ":", "")
	
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.macvendors.com/" + cleanOUI)
	if err != nil {
		return VendorInfo{}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return VendorInfo{}
	}
	
	var buf [256]byte
	n, _ := resp.Body.Read(buf[:])
	manufacturer := strings.TrimSpace(string(buf[:n]))
	
	if manufacturer == "" {
		return VendorInfo{}
	}
	
	return VendorInfo{Manufacturer: manufacturer}
}

// DetectISPFromSSID tries to identify ISP from SSID pattern
func DetectISPFromSSID(ssid string) string {
	ssidLower := strings.ToLower(ssid)
	
	patterns := map[string]*regexp.Regexp{
		"Free":     regexp.MustCompile(`freebox|free[-_]?wifi|free[-_]?\d`),
		"Orange":   regexp.MustCompile(`livebox|orange[-_]?\d|orange[-_]?wifi`),
		"SFR":      regexp.MustCompile(`sfr[-_]?box|sfr[-_]?\d|sfr[-_]?wifi|neufbox`),
		"Bouygues": regexp.MustCompile(`bbox|bouygues|bytel`),
	}
	
	for isp, pattern := range patterns {
		if pattern.MatchString(ssidLower) {
			return isp
		}
	}
	
	return ""
}

// GetPasswordPattern returns the password pattern for a given ISP
func GetPasswordPattern(isp string) (PasswordPattern, bool) {
	pattern, ok := passwordPatterns[isp]
	return pattern, ok
}

// AnalyzeNetwork returns full analysis of a network including vendor and password pattern
func AnalyzeNetwork(bssid, ssid string) map[string]any {
	result := map[string]any{
		"bssid": bssid,
		"ssid":  ssid,
	}
	
	// Get vendor from BSSID
	vendor := LookupVendor(bssid)
	result["vendor"] = vendor
	
	// Try to detect ISP
	isp := vendor.ISP
	if isp == "" {
		isp = DetectISPFromSSID(ssid)
	}
	result["isp"] = isp
	
	// Get password pattern if ISP is known
	if isp != "" {
		if pattern, ok := GetPasswordPattern(isp); ok {
			result["password_pattern"] = pattern
			result["has_pattern"] = true
		} else {
			result["has_pattern"] = false
		}
	}
	
	return result
}

// GenerateWordlistSuggestion suggests how to generate a targeted wordlist
func GenerateWordlistSuggestion(isp string) map[string]any {
	pattern, ok := GetPasswordPattern(isp)
	if !ok {
		return map[string]any{
			"available": false,
			"message":   "Pattern inconnu pour cet ISP",
		}
	}
	
	suggestion := map[string]any{
		"available":   true,
		"isp":         isp,
		"pattern":     pattern,
		"description": pattern.Description,
	}
	
	switch pattern.GeneratorType {
	case "mask":
		suggestion["tool"] = "hashcat"
		suggestion["command"] = fmt.Sprintf("hashcat -m 22000 -a 3 capture.hc22000 '%s'", pattern.MaskPattern)
		suggestion["note"] = "Utilisez hashcat avec un masque pour une attaque ciblée"
		
	case "wordlist-combo":
		suggestion["tool"] = "custom"
		suggestion["command"] = "Wordlist de mots latins + règles hashcat"
		suggestion["note"] = "Pour Freebox: combiner dictionnaire latin avec suffixes"
		
	default:
		suggestion["tool"] = "aircrack-ng"
		suggestion["command"] = "aircrack-ng -w wordlist.txt capture.cap"
	}
	
	return suggestion
}

// SaveVendorCache saves vendor cache to a file
func SaveVendorCache(path string) error {
	data, err := json.MarshalIndent(knownOUI, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadVendorCache loads vendor cache from a file
func LoadVendorCache(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &knownOUI)
}
