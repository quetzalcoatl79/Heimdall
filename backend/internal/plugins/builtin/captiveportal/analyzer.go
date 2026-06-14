package captiveportal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// FormField represents a detected form field
type FormField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	ID          string `json:"id,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required"`
	MaxLength   int    `json:"maxLength,omitempty"`
	MinLength   int    `json:"minLength,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
}

// FormAnalysis contains the analysis result of a captive portal page
type FormAnalysis struct {
	URL           string      `json:"url"`
	Title         string      `json:"title"`
	FormAction    string      `json:"form_action"`
	FormMethod    string      `json:"form_method"`
	Fields        []FormField `json:"fields"`
	HasCSRF       bool        `json:"has_csrf"`
	CSRFFieldName string      `json:"csrf_field_name,omitempty"`
	HasCaptcha    bool        `json:"has_captcha"`
	AnalyzedAt    time.Time   `json:"analyzed_at"`
	RawHTML       string      `json:"raw_html,omitempty"`
	Error         string      `json:"error,omitempty"`
}

// AnalyzePortalPage fetches and analyzes a captive portal page
func AnalyzePortalPage(url string) FormAnalysis {
	analysis := FormAnalysis{
		URL:        url,
		AnalyzedAt: time.Now(),
		Fields:     []FormField{},
	}

	// Fetch the page
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects but track them
			return nil
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		analysis.Error = fmt.Sprintf("Erreur de connexion: %v", err)
		return analysis
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		analysis.Error = fmt.Sprintf("Erreur de lecture: %v", err)
		return analysis
	}

	html := string(body)
	analysis.RawHTML = html

	// Extract title
	titleRe := regexp.MustCompile(`<title>([^<]*)</title>`)
	if matches := titleRe.FindStringSubmatch(html); len(matches) > 1 {
		analysis.Title = strings.TrimSpace(matches[1])
	}

	// Find forms
	formRe := regexp.MustCompile(`(?is)<form[^>]*action=["']([^"']*)["'][^>]*method=["']([^"']*)["'][^>]*>`)
	formReAlt := regexp.MustCompile(`(?is)<form[^>]*method=["']([^"']*)["'][^>]*action=["']([^"']*)["'][^>]*>`)

	if matches := formRe.FindStringSubmatch(html); len(matches) > 2 {
		analysis.FormAction = matches[1]
		analysis.FormMethod = strings.ToUpper(matches[2])
	} else if matches := formReAlt.FindStringSubmatch(html); len(matches) > 2 {
		analysis.FormMethod = strings.ToUpper(matches[1])
		analysis.FormAction = matches[2]
	}

	// Extract input fields
	inputRe := regexp.MustCompile(`(?is)<input[^>]*>`)
	inputs := inputRe.FindAllString(html, -1)

	for _, input := range inputs {
		field := parseInputField(input)
		if field.Name != "" || field.ID != "" {
			analysis.Fields = append(analysis.Fields, field)

			// Detect CSRF tokens
			nameLower := strings.ToLower(field.Name)
			if strings.Contains(nameLower, "csrf") || strings.Contains(nameLower, "token") ||
				strings.Contains(nameLower, "_token") || field.Name == "__RequestVerificationToken" {
				analysis.HasCSRF = true
				analysis.CSRFFieldName = field.Name
			}
		}
	}

	// Detect captcha
	captchaPatterns := []string{
		"captcha", "recaptcha", "g-recaptcha", "h-captcha", "cf-turnstile",
	}
	htmlLower := strings.ToLower(html)
	for _, pattern := range captchaPatterns {
		if strings.Contains(htmlLower, pattern) {
			analysis.HasCaptcha = true
			break
		}
	}

	return analysis
}

// parseInputField extracts field information from an input HTML tag
func parseInputField(input string) FormField {
	field := FormField{}

	// Extract name
	nameRe := regexp.MustCompile(`name=["']([^"']*)["']`)
	if matches := nameRe.FindStringSubmatch(input); len(matches) > 1 {
		field.Name = matches[1]
	}

	// Extract type
	typeRe := regexp.MustCompile(`type=["']([^"']*)["']`)
	if matches := typeRe.FindStringSubmatch(input); len(matches) > 1 {
		field.Type = matches[1]
	} else {
		field.Type = "text" // default
	}

	// Extract ID
	idRe := regexp.MustCompile(`id=["']([^"']*)["']`)
	if matches := idRe.FindStringSubmatch(input); len(matches) > 1 {
		field.ID = matches[1]
	}

	// Extract placeholder
	placeholderRe := regexp.MustCompile(`placeholder=["']([^"']*)["']`)
	if matches := placeholderRe.FindStringSubmatch(input); len(matches) > 1 {
		field.Placeholder = matches[1]
	}

	// Check required
	field.Required = strings.Contains(input, "required")

	// Extract maxlength
	maxLengthRe := regexp.MustCompile(`maxlength=["']?(\d+)["']?`)
	if matches := maxLengthRe.FindStringSubmatch(input); len(matches) > 1 {
		fmt.Sscanf(matches[1], "%d", &field.MaxLength)
	}

	// Extract minlength
	minLengthRe := regexp.MustCompile(`minlength=["']?(\d+)["']?`)
	if matches := minLengthRe.FindStringSubmatch(input); len(matches) > 1 {
		fmt.Sscanf(matches[1], "%d", &field.MinLength)
	}

	// Extract pattern
	patternRe := regexp.MustCompile(`pattern=["']([^"']*)["']`)
	if matches := patternRe.FindStringSubmatch(input); len(matches) > 1 {
		field.Pattern = matches[1]
	}

	return field
}
