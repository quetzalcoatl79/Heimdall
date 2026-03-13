package wifi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

// ReportConfig holds configuration for report generation
type ReportConfig struct {
	AuditID     uuid.UUID `json:"audit_id"`
	ClientName  string    `json:"client_name"`
	TesterName  string    `json:"tester_name,omitempty"`
	CompanyLogo string    `json:"company_logo,omitempty"` // Path to logo image
	OutputPath  string    `json:"output_path,omitempty"`
	IncludeRaw  bool      `json:"include_raw_data,omitempty"`
}

// ReportData holds all data needed for report generation
type ReportData struct {
	// Client info
	ClientName    string `json:"client_name"`
	ClientContact string `json:"client_contact,omitempty"`
	Location      string `json:"location,omitempty"`
	// Audit info
	AuditID    string    `json:"audit_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	TesterName string    `json:"tester_name"`
	// Summary
	NetworksScanned    int `json:"networks_scanned"`
	NetworksTested     int `json:"networks_tested"`
	HandshakesCaptured int `json:"handshakes_captured"`
	PasswordsCracked   int `json:"passwords_cracked"`
	// Details
	Networks []NetworkReportEntry `json:"networks"`
	// Recommendations
	GlobalRecommendations []string `json:"global_recommendations"`
}

// NetworkReportEntry contains details about a tested network
type NetworkReportEntry struct {
	SSID               string    `json:"ssid"`
	BSSID              string    `json:"bssid"`
	Channel            int       `json:"channel"`
	Security           string    `json:"security"`
	Vendor             string    `json:"vendor,omitempty"`
	ISP                string    `json:"isp,omitempty"`
	HandshakeCaptured  bool      `json:"handshake_captured"`
	PasswordCracked    bool      `json:"password_cracked"`
	Password           string    `json:"password,omitempty"`
	CrackMethod        string    `json:"crack_method,omitempty"`
	CrackDuration      float64   `json:"crack_duration_seconds,omitempty"`
	VulnerabilityLevel string    `json:"vulnerability_level"`
	Recommendations    []string  `json:"recommendations"`
	TestedAt           time.Time `json:"tested_at"`
}

// Color definitions for PDF
type rgbColor struct{ R, G, B int }

var (
	colorPrimary = rgbColor{41, 128, 185}  // Blue
	colorDanger  = rgbColor{192, 57, 43}   // Red
	colorSuccess = rgbColor{39, 174, 96}   // Green
	colorWarning = rgbColor{243, 156, 18}  // Orange
	colorDark    = rgbColor{44, 62, 80}    // Dark blue
	colorLight   = rgbColor{236, 240, 241} // Light gray
	colorWhite   = rgbColor{255, 255, 255} // White
)

// GenerateReport creates a PDF report from audit data
func GenerateReport(data ReportData, config ReportConfig) (string, error) {
	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		reportsDir := "/opt/heimdall/reports"
		if err := os.MkdirAll(reportsDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create reports directory: %v", err)
		}
		timestamp := time.Now().Format("20060102_150405")
		safeClientName := strings.ReplaceAll(data.ClientName, " ", "_")
		safeClientName = strings.ReplaceAll(safeClientName, "/", "-")
		outputPath = filepath.Join(reportsDir, fmt.Sprintf("rapport_wifi_%s_%s.pdf", safeClientName, timestamp))
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 20)

	// Add fonts
	pdf.AddPage()

	// Title page
	addTitlePage(pdf, data, config)

	// Executive summary
	pdf.AddPage()
	addExecutiveSummary(pdf, data)

	// Network details
	if len(data.Networks) > 0 {
		pdf.AddPage()
		addNetworkDetails(pdf, data)
	}

	// Recommendations
	pdf.AddPage()
	addRecommendations(pdf, data)

	// Footer with page numbers
	addPageNumbers(pdf)

	// Save PDF
	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return "", fmt.Errorf("failed to save PDF: %v", err)
	}

	return outputPath, nil
}

func addTitlePage(pdf *gofpdf.Fpdf, data ReportData, config ReportConfig) {
	// Logo if provided
	if config.CompanyLogo != "" {
		if _, err := os.Stat(config.CompanyLogo); err == nil {
			pdf.ImageOptions(config.CompanyLogo, 80, 20, 50, 0, false, gofpdf.ImageOptions{}, 0, "")
			pdf.Ln(60)
		}
	} else {
		pdf.Ln(40)
	}

	// Title
	pdf.SetFont("Helvetica", "B", 28)
	pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)
	pdf.CellFormat(0, 15, "RAPPORT D'AUDIT", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 24)
	pdf.SetTextColor(colorPrimary.R, colorPrimary.G, colorPrimary.B)
	pdf.CellFormat(0, 12, "SECURITE WIFI", "", 1, "C", false, 0, "")

	pdf.Ln(10)

	// Decorative line
	pdf.SetDrawColor(colorPrimary.R, colorPrimary.G, colorPrimary.B)
	pdf.SetLineWidth(1)
	pdf.Line(60, pdf.GetY(), 150, pdf.GetY())
	pdf.Ln(15)

	// Client info
	pdf.SetFont("Helvetica", "", 14)
	pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)
	pdf.CellFormat(0, 10, fmt.Sprintf("Client: %s", data.ClientName), "", 1, "C", false, 0, "")

	if data.Location != "" {
		pdf.SetFont("Helvetica", "", 12)
		pdf.CellFormat(0, 8, fmt.Sprintf("Lieu: %s", data.Location), "", 1, "C", false, 0, "")
	}

	pdf.Ln(5)

	// Dates
	pdf.SetFont("Helvetica", "", 12)
	dateFormat := "02/01/2006"
	pdf.CellFormat(0, 8, fmt.Sprintf("Date de debut: %s", data.StartDate.Format(dateFormat)), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Date de fin: %s", data.EndDate.Format(dateFormat)), "", 1, "C", false, 0, "")

	pdf.Ln(10)

	// Tester
	if data.TesterName != "" {
		pdf.SetFont("Helvetica", "I", 11)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(0, 8, fmt.Sprintf("Auditeur: %s", data.TesterName), "", 1, "C", false, 0, "")
	}

	// Classification
	pdf.Ln(30)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(colorDanger.R, colorDanger.G, colorDanger.B)
	pdf.CellFormat(0, 8, "CONFIDENTIEL", "", 1, "C", false, 0, "")

	// Footer with generation date
	pdf.SetY(-30)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 5, fmt.Sprintf("Rapport genere le %s", time.Now().Format("02/01/2006 a 15:04")), "", 1, "C", false, 0, "")
}

func addExecutiveSummary(pdf *gofpdf.Fpdf, data ReportData) {
	// Section title
	addSectionTitle(pdf, "RESUME EXECUTIF")

	pdf.Ln(5)

	// Summary box
	pdf.SetFillColor(colorLight.R, colorLight.G, colorLight.B)
	pdf.Rect(15, pdf.GetY(), 180, 50, "F")

	pdf.SetY(pdf.GetY() + 5)
	pdf.SetX(20)

	// Stats in columns
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)

	// Row 1
	pdf.SetX(25)
	pdf.CellFormat(80, 10, fmt.Sprintf("Reseaux scannes: %d", data.NetworksScanned), "", 0, "L", false, 0, "")
	pdf.CellFormat(80, 10, fmt.Sprintf("Reseaux testes: %d", data.NetworksTested), "", 1, "L", false, 0, "")

	// Row 2
	pdf.SetX(25)
	pdf.CellFormat(80, 10, fmt.Sprintf("Handshakes captures: %d", data.HandshakesCaptured), "", 0, "L", false, 0, "")

	// Password cracked - highlight in red/green
	if data.PasswordsCracked > 0 {
		pdf.SetTextColor(colorDanger.R, colorDanger.G, colorDanger.B)
	} else {
		pdf.SetTextColor(colorSuccess.R, colorSuccess.G, colorSuccess.B)
	}
	pdf.CellFormat(80, 10, fmt.Sprintf("Mots de passe trouves: %d", data.PasswordsCracked), "", 1, "L", false, 0, "")

	pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)

	// Security score
	pdf.Ln(10)
	securityScore := calculateSecurityScore(data)
	scoreColor := getScoreColor(securityScore)

	pdf.SetX(25)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(40, 10, "Score de securite:", "", 0, "L", false, 0, "")
	pdf.SetTextColor(scoreColor.R, scoreColor.G, scoreColor.B)
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(30, 10, fmt.Sprintf("%d/100", securityScore), "", 0, "L", false, 0, "")

	pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 10, getScoreLabel(securityScore), "", 1, "L", false, 0, "")

	pdf.Ln(15)

	// Risk summary
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Analyse des risques:", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 11)
	risks := analyzeRisks(data)
	for _, risk := range risks {
		pdf.SetX(20)
		pdf.MultiCell(0, 6, fmt.Sprintf("- %s", risk), "", "L", false)
	}
}

func addNetworkDetails(pdf *gofpdf.Fpdf, data ReportData) {
	addSectionTitle(pdf, "DETAILS DES RESEAUX TESTES")

	pdf.Ln(5)

	for i, network := range data.Networks {
		// Check if we need a new page
		if pdf.GetY() > 240 {
			pdf.AddPage()
		}

		// Network header
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetFillColor(colorPrimary.R, colorPrimary.G, colorPrimary.B)
		pdf.SetTextColor(colorWhite.R, colorWhite.G, colorWhite.B)
		pdf.CellFormat(0, 8, fmt.Sprintf(" %d. %s", i+1, network.SSID), "", 1, "L", true, 0, "")

		pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)
		pdf.SetFont("Helvetica", "", 10)

		// Network info table
		pdf.SetX(20)
		pdf.CellFormat(40, 6, "BSSID:", "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, network.BSSID, "", 1, "L", false, 0, "")

		pdf.SetX(20)
		pdf.CellFormat(40, 6, "Canal:", "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%d", network.Channel), "", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, "Securite:", "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, network.Security, "", 1, "L", false, 0, "")

		if network.Vendor != "" || network.ISP != "" {
			pdf.SetX(20)
			pdf.CellFormat(40, 6, "Fabricant/ISP:", "", 0, "L", false, 0, "")
			vendor := network.Vendor
			if network.ISP != "" {
				vendor = fmt.Sprintf("%s (%s)", network.Vendor, network.ISP)
			}
			pdf.CellFormat(0, 6, vendor, "", 1, "L", false, 0, "")
		}

		// Results
		pdf.SetX(20)
		pdf.CellFormat(40, 6, "Handshake:", "", 0, "L", false, 0, "")
		if network.HandshakeCaptured {
			pdf.SetTextColor(colorSuccess.R, colorSuccess.G, colorSuccess.B)
			pdf.CellFormat(0, 6, "Capture", "", 1, "L", false, 0, "")
		} else {
			pdf.SetTextColor(100, 100, 100)
			pdf.CellFormat(0, 6, "Non capture", "", 1, "L", false, 0, "")
		}
		pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)

		// Password result - CRITICAL
		if network.PasswordCracked {
			pdf.SetX(20)
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetTextColor(colorDanger.R, colorDanger.G, colorDanger.B)
			pdf.CellFormat(40, 6, "MOT DE PASSE:", "", 0, "L", false, 0, "")
			pdf.CellFormat(0, 6, network.Password, "", 1, "L", false, 0, "")

			if network.CrackMethod != "" {
				pdf.SetX(20)
				pdf.SetFont("Helvetica", "", 9)
				pdf.SetTextColor(100, 100, 100)
				pdf.CellFormat(40, 6, "Methode:", "", 0, "L", false, 0, "")
				method := network.CrackMethod
				if network.CrackDuration > 0 {
					method += fmt.Sprintf(" (%.1fs)", network.CrackDuration)
				}
				pdf.CellFormat(0, 6, method, "", 1, "L", false, 0, "")
			}

			pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)
		}

		// Vulnerability level
		pdf.SetX(20)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(40, 6, "Vulnerabilite:", "", 0, "L", false, 0, "")
		vulnColor := getVulnerabilityColor(network.VulnerabilityLevel)
		pdf.SetTextColor(vulnColor.R, vulnColor.G, vulnColor.B)
		pdf.CellFormat(0, 6, strings.ToUpper(network.VulnerabilityLevel), "", 1, "L", false, 0, "")
		pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)

		// Network recommendations
		if len(network.Recommendations) > 0 {
			pdf.SetFont("Helvetica", "I", 9)
			pdf.SetX(20)
			pdf.CellFormat(0, 5, "Recommandations:", "", 1, "L", false, 0, "")
			for _, rec := range network.Recommendations {
				pdf.SetX(25)
				pdf.MultiCell(165, 4, fmt.Sprintf("- %s", rec), "", "L", false)
			}
		}

		pdf.Ln(5)
	}
}

func addRecommendations(pdf *gofpdf.Fpdf, data ReportData) {
	addSectionTitle(pdf, "RECOMMANDATIONS GLOBALES")

	pdf.Ln(5)

	recommendations := data.GlobalRecommendations
	if len(recommendations) == 0 {
		recommendations = generateDefaultRecommendations(data)
	}

	pdf.SetFont("Helvetica", "", 11)

	for i, rec := range recommendations {
		if pdf.GetY() > 260 {
			pdf.AddPage()
		}

		// Numbered recommendation
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(colorPrimary.R, colorPrimary.G, colorPrimary.B)
		pdf.CellFormat(10, 7, fmt.Sprintf("%d.", i+1), "", 0, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(colorDark.R, colorDark.G, colorDark.B)
		pdf.MultiCell(170, 7, rec, "", "L", false)
		pdf.Ln(2)
	}

	// Signature section
	pdf.Ln(20)
	pdf.SetDrawColor(colorDark.R, colorDark.G, colorDark.B)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "I", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 6, "Ce rapport est confidentiel et destine uniquement au client mentionne.", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 6, "Toute reproduction ou diffusion non autorisee est interdite.", "", 1, "C", false, 0, "")
}

func addSectionTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(colorPrimary.R, colorPrimary.G, colorPrimary.B)
	pdf.CellFormat(0, 10, title, "", 1, "L", false, 0, "")

	pdf.SetDrawColor(colorPrimary.R, colorPrimary.G, colorPrimary.B)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 80, pdf.GetY())
	pdf.Ln(3)
}

func addPageNumbers(pdf *gofpdf.Fpdf) {
	pdf.AliasNbPages("")
	// Page numbers would be added via footer callback, but gofpdf handles this differently
	// For now, page numbers are implicit in the PDF viewer
}

// Helper functions

func calculateSecurityScore(data ReportData) int {
	if data.NetworksTested == 0 {
		return 100
	}

	// Start with 100, subtract for vulnerabilities
	score := 100

	// Major penalty for cracked passwords
	crackedRatio := float64(data.PasswordsCracked) / float64(data.NetworksTested)
	score -= int(crackedRatio * 60)

	// Penalty for captured handshakes (even if not cracked)
	handshakeRatio := float64(data.HandshakesCaptured) / float64(data.NetworksTested)
	score -= int(handshakeRatio * 20)

	if score < 0 {
		score = 0
	}

	return score
}

func getScoreColor(score int) rgbColor {
	if score >= 80 {
		return colorSuccess
	} else if score >= 50 {
		return colorWarning
	}
	return colorDanger
}

func getScoreLabel(score int) string {
	if score >= 90 {
		return "Excellent"
	} else if score >= 80 {
		return "Bon"
	} else if score >= 60 {
		return "Moyen"
	} else if score >= 40 {
		return "Faible"
	}
	return "Critique"
}

func getVulnerabilityColor(level string) rgbColor {
	switch strings.ToLower(level) {
	case "critical":
		return colorDanger
	case "high":
		return rgbColor{231, 76, 60}
	case "medium":
		return colorWarning
	case "low":
		return colorSuccess
	default:
		return rgbColor{100, 100, 100}
	}
}

func analyzeRisks(data ReportData) []string {
	var risks []string

	if data.PasswordsCracked > 0 {
		risks = append(risks, fmt.Sprintf("CRITIQUE: %d mot(s) de passe WiFi ont ete compromis.", data.PasswordsCracked))
	}

	if data.HandshakesCaptured > data.PasswordsCracked {
		risks = append(risks, fmt.Sprintf("ELEVE: %d handshake(s) capture(s) pourraient etre crackes avec plus de ressources.",
			data.HandshakesCaptured-data.PasswordsCracked))
	}

	// Check for weak security protocols
	weakSecurityCount := 0
	for _, net := range data.Networks {
		if strings.Contains(strings.ToLower(net.Security), "wep") {
			weakSecurityCount++
		}
	}
	if weakSecurityCount > 0 {
		risks = append(risks, fmt.Sprintf("CRITIQUE: %d reseau(x) utilisent le protocole WEP obsolete.", weakSecurityCount))
	}

	if len(risks) == 0 {
		risks = append(risks, "Aucune vulnerabilite majeure detectee lors de cet audit.")
	}

	return risks
}

func generateDefaultRecommendations(data ReportData) []string {
	var recs []string

	// Always recommended
	recs = append(recs, "Utiliser WPA3 ou WPA2-Enterprise pour tous les reseaux WiFi professionnels.")
	recs = append(recs, "Choisir des mots de passe WiFi d'au moins 16 caracteres avec melange de lettres, chiffres et symboles.")

	if data.PasswordsCracked > 0 {
		recs = append(recs, "URGENT: Changer immediatement les mots de passe WiFi compromis.")
		recs = append(recs, "Verifier si des donnees sensibles ont pu etre interceptees via les reseaux compromis.")
	}

	if data.HandshakesCaptured > 0 {
		recs = append(recs, "Activer la detection d'intrusion (IDS) pour surveiller les tentatives de deauthentification.")
		recs = append(recs, "Considerer l'utilisation de 802.11w (Management Frame Protection) pour prevenir les attaques de deauth.")
	}

	recs = append(recs, "Former les employes aux bonnes pratiques de securite WiFi.")
	recs = append(recs, "Effectuer des audits de securite WiFi reguliers (au moins annuellement).")

	return recs
}

// DetermineVulnerabilityLevel determines the vulnerability level of a network
func DetermineVulnerabilityLevel(network NetworkReportEntry) string {
	if network.PasswordCracked {
		return "critical"
	}
	if network.HandshakeCaptured {
		if strings.Contains(strings.ToLower(network.Security), "wep") {
			return "critical"
		}
		return "high"
	}
	if strings.Contains(strings.ToLower(network.Security), "wep") {
		return "critical"
	}
	if strings.Contains(strings.ToLower(network.Security), "wpa3") {
		return "low"
	}
	return "medium"
}
