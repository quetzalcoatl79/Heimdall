package wifi

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WordlistInfo represents a known wordlist that can be downloaded
type WordlistInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Size        string `json:"size"`        // Human readable size
	Compressed  bool   `json:"compressed"`  // Is it gzipped?
	Priority    int    `json:"priority"`    // Higher = try first for attacks
	Category    string `json:"category"`    // wifi, general, french, latin
	Installed   bool   `json:"installed"`   // Whether it's downloaded
	LocalPath   string `json:"local_path"`  // Path if installed
	LocalSize   int64  `json:"local_size,omitempty"`  // Bytes if installed
	LocalLines  int64  `json:"local_lines,omitempty"` // Lines if installed
}

// WordlistsDir is the default directory for wordlists
const WordlistsDir = "/opt/heimdall/wordlists"

// KnownWordlists contains popular wordlists useful for WiFi cracking
var KnownWordlists = []WordlistInfo{
	// ===== MOST USEFUL FOR WIFI =====
	{
		Name:        "rockyou",
		Description: "Le plus célèbre: 14 millions de mots de passe réels (2009 breach)",
		URL:         "https://github.com/brannondorsey/naive-hashcat/releases/download/data/rockyou.txt",
		Size:        "139 MB",
		Compressed:  false,
		Priority:    100,
		Category:    "wifi",
	},
	{
		Name:        "probable-wpa",
		Description: "Wordlist optimisée WPA (8-63 chars) - triée par probabilité",
		URL:         "https://github.com/berzerk0/Probable-Wordlists/raw/master/Real-Passwords/WPA-Length/Top62Million-probable-WPA.txt.gz",
		Size:        "320 MB (gzipped)",
		Compressed:  true,
		Priority:    95,
		Category:    "wifi",
	},
	{
		Name:        "top1m-probable",
		Description: "Top 1 million de mots de passe les plus courants",
		URL:         "https://github.com/berzerk0/Probable-Wordlists/raw/master/Real-Passwords/Top1575-probable-v2.txt",
		Size:        "17 KB",
		Compressed:  false,
		Priority:    90,
		Category:    "general",
	},
	{
		Name:        "crackstation-human",
		Description: "Seulement des mots de passe humains (15GB complet)",
		URL:         "https://crackstation.net/files/crackstation-human-only.txt.gz",
		Size:        "684 MB (gzipped)",
		Compressed:  true,
		Priority:    80,
		Category:    "general",
	},
	
	// ===== FRENCH SPECIFIC =====
	{
		Name:        "french-passwords",
		Description: "Mots de passe français courants",
		URL:         "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Common-Credentials/Language-Specific/French/french-passwords-10k.txt",
		Size:        "115 KB",
		Compressed:  false,
		Priority:    85,
		Category:    "french",
	},
	
	// ===== LATIN (Freebox) =====
	{
		Name:        "latin-heimdall",
		Description: "Mots latins Freebox (intégré à Heimdall)",
		URL:         "",  // Generated locally
		Size:        "~50 KB",
		Compressed:  false,
		Priority:    70,
		Category:    "latin",
	},
	
	// ===== ROUTER DEFAULTS =====
	{
		Name:        "default-passwords",
		Description: "Mots de passe par défaut de routeurs",
		URL:         "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/Default-Credentials/default-passwords.txt",
		Size:        "24 KB",
		Compressed:  false,
		Priority:    60,
		Category:    "router",
	},
	{
		Name:        "wifi-common",
		Description: "Mots de passe WiFi communs (SecLists)",
		URL:         "https://raw.githubusercontent.com/danielmiessler/SecLists/master/Passwords/WiFi-WPA/wifi-common.txt",
		Size:        "~5 KB",
		Compressed:  false,
		Priority:    55,
		Category:    "wifi",
	},
	
	// ===== NUMBERS/PATTERNS =====
	{
		Name:        "phone-numbers-fr",
		Description: "Numéros de téléphone français (06XXXXXXXX, etc.)",
		URL:         "",  // Generated locally
		Size:        "~100 MB (generated)",
		Compressed:  false,
		Priority:    50,
		Category:    "french",
	},
}

// GetWordlistsDir ensures the wordlists directory exists
func GetWordlistsDir() string {
	// 1) Explicit override
	if customDir := strings.TrimSpace(os.Getenv("WORDLISTS_DIR")); customDir != "" {
		if !filepath.IsAbs(customDir) {
			if cwd, err := os.Getwd(); err == nil {
				customDir = filepath.Join(cwd, customDir)
			}
		}
		if err := os.MkdirAll(customDir, 0755); err == nil {
			return customDir
		}
	}

	// 2) App root override
	if root := strings.TrimSpace(os.Getenv("HEIMDALL_ROOT")); root != "" {
		if !filepath.IsAbs(root) {
			if cwd, err := os.Getwd(); err == nil {
				root = filepath.Join(cwd, root)
			}
		}
		dir := filepath.Join(root, "wordlists")
		if err := os.MkdirAll(dir, 0755); err == nil {
			return dir
		}
	}

	// 3) Default system location
	if err := os.MkdirAll(WordlistsDir, 0755); err == nil {
		return WordlistsDir
	}

	// 4) Fallback to /tmp if /opt not writable
	tmpDir := filepath.Join(os.TempDir(), "heimdall-wordlists")
	_ = os.MkdirAll(tmpDir, 0755)
	return tmpDir
}

// ListWordlists returns all available wordlists with their installation status
func ListWordlists() []WordlistInfo {
	dir := GetWordlistsDir()
	result := make([]WordlistInfo, len(KnownWordlists))
	copy(result, KnownWordlists)
	
	for i := range result {
		// Check if installed
		localPath := filepath.Join(dir, result[i].Name+".txt")
		if _, err := os.Stat(localPath); err == nil {
			result[i].Installed = true
			result[i].LocalPath = localPath
			if info, err := os.Stat(localPath); err == nil {
				result[i].LocalSize = info.Size()
				if info.Size() < 200*1024*1024 {
					if data, err := exec.Command("wc", "-l", localPath).Output(); err == nil {
						fmt.Sscanf(string(data), "%d", &result[i].LocalLines)
					}
				} else {
					result[i].LocalLines = info.Size() / 10
				}
			}
		}
		
		// Special case for latin-heimdall (always available)
		if result[i].Name == "latin-heimdall" {
			localPath := GetFreeboxWordlistPath()
			if localPath != "" {
				result[i].Installed = true
				result[i].LocalPath = localPath
				if info, err := os.Stat(localPath); err == nil {
					result[i].LocalSize = info.Size()
					if info.Size() < 200*1024*1024 {
						if data, err := exec.Command("wc", "-l", localPath).Output(); err == nil {
							fmt.Sscanf(string(data), "%d", &result[i].LocalLines)
						}
					} else {
						result[i].LocalLines = info.Size() / 10
					}
				}
			}
		}
	}
	
	return result
}

// DownloadWordlist downloads a wordlist from the internet
func DownloadWordlist(name string) (string, error) {
	// Find the wordlist
	var wl *WordlistInfo
	for _, w := range KnownWordlists {
		if strings.EqualFold(w.Name, name) {
			wl = &w
			break
		}
	}
	
	if wl == nil {
		return "", fmt.Errorf("wordlist %s non trouvée", name)
	}
	
	// Special cases for locally generated lists
	if wl.Name == "latin-heimdall" {
		return GetFreeboxWordlistPath(), nil
	}
	
	if wl.Name == "phone-numbers-fr" {
		return generateFrenchPhoneNumbers()
	}
	
	if wl.URL == "" {
		return "", fmt.Errorf("wordlist %s ne peut pas être téléchargée", name)
	}
	
	dir := GetWordlistsDir()
	localPath := filepath.Join(dir, wl.Name+".txt")
	
	// Download with progress
	fmt.Printf("[WORDLIST] 📥 Téléchargement de %s...\n", wl.Name)
	
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Get(wl.URL)
	if err != nil {
		return "", fmt.Errorf("erreur téléchargement: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur HTTP %d", resp.StatusCode)
	}
	
	// Create output file
	outFile, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("erreur création fichier: %v", err)
	}
	defer outFile.Close()
	
	var reader io.Reader = resp.Body
	
	// Handle gzipped files
	if wl.Compressed {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("erreur décompression: %v", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}
	
	written, err := io.Copy(outFile, reader)
	if err != nil {
		os.Remove(localPath)
		return "", fmt.Errorf("erreur écriture: %v", err)
	}
	
	fmt.Printf("[WORDLIST] ✅ %s téléchargé (%d bytes)\n", wl.Name, written)
	return localPath, nil
}

// GetBestWordlistsForISP returns the best wordlists for a given ISP
func GetBestWordlistsForISP(isp string) []string {
	available := ListWordlists()
	var result []string
	
	// Prioritize by ISP
	switch strings.ToLower(isp) {
	case "free", "freebox":
		// For Freebox, prefer Latin wordlist first
		for _, wl := range available {
			if wl.Installed && wl.Category == "latin" {
				result = append(result, wl.LocalPath)
			}
		}
	case "orange", "livebox":
		// Orange often uses patterns like adjective+noun+digits
		// Start with french-passwords
		for _, wl := range available {
			if wl.Installed && wl.Category == "french" {
				result = append(result, wl.LocalPath)
			}
		}
	case "sfr", "neufbox":
		// SFR uses hex patterns - less useful with wordlists
		// Still try common ones
	case "bouygues", "bbox":
		// Bouygues uses alphanumeric patterns
	}
	
	// Add general wordlists sorted by priority
	type priorityPath struct {
		priority int
		path     string
	}
	var general []priorityPath
	
	for _, wl := range available {
		if wl.Installed && wl.Category != "latin" {
			general = append(general, priorityPath{wl.Priority, wl.LocalPath})
		}
	}
	
	// Sort by priority (higher first)
	for i := 0; i < len(general); i++ {
		for j := i + 1; j < len(general); j++ {
			if general[j].priority > general[i].priority {
				general[i], general[j] = general[j], general[i]
			}
		}
	}
	
	for _, gp := range general {
		// Don't add duplicates
		found := false
		for _, r := range result {
			if r == gp.path {
				found = true
				break
			}
		}
		if !found {
			result = append(result, gp.path)
		}
	}
	
	return result
}

// generateFrenchPhoneNumbers creates a wordlist of French phone numbers
func generateFrenchPhoneNumbers() (string, error) {
	dir := GetWordlistsDir()
	localPath := filepath.Join(dir, "phone-numbers-fr.txt")
	
	// Check if already exists
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}
	
	fmt.Println("[WORDLIST] 📱 Génération des numéros de téléphone français...")
	
	file, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	// Generate common patterns
	// 06XXXXXXXX (mobile) - most common as passwords
	// 07XXXXXXXX (mobile)
	
	// Only generate a subset to avoid huge files
	// Common patterns: 0601020304, 0612345678, etc.
	
	prefixes := []string{"06", "07"}
	count := 0
	
	for _, prefix := range prefixes {
		// Generate common patterns
		// XXXXXX00 to XXXXXX99 for common suffixes
		for i := 0; i < 100000000; i += 1000 { // Skip to reduce size
			number := fmt.Sprintf("%s%08d", prefix, i)
			fmt.Fprintln(file, number)
			count++
			if count > 1000000 { // Limit to 1M entries
				break
			}
		}
		if count > 1000000 {
			break
		}
	}
	
	fmt.Printf("[WORDLIST] ✅ phone-numbers-fr généré (%d numéros)\n", count)
	return localPath, nil
}

// EnsureRockyou downloads rockyou if not present (most important wordlist)
func EnsureRockyou() string {
	dir := GetWordlistsDir()
	localPath := filepath.Join(dir, "rockyou.txt")
	
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	
	// Also check system locations
	systemPaths := []string{
		"/usr/share/wordlists/rockyou.txt",
		"/usr/share/john/password.lst",
		"/opt/wordlists/rockyou.txt",
	}
	
	for _, p := range systemPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	
	// Try to download
	path, err := DownloadWordlist("rockyou")
	if err != nil {
		fmt.Printf("[WORDLIST] ⚠️ Impossible de télécharger rockyou: %v\n", err)
		return ""
	}
	
	return path
}

// GetInstalledWordlists returns paths to all installed wordlists
func GetInstalledWordlists() []string {
	available := ListWordlists()
	var result []string
	
	for _, wl := range available {
		if wl.Installed {
			result = append(result, wl.LocalPath)
		}
	}
	
	return result
}
