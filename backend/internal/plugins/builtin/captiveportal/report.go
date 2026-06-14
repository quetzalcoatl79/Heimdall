package captiveportal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// CaptivePortalReportConfig holds configuration for report generation
type CaptivePortalReportConfig struct {
	ClientName  string `json:"client_name"`
	TesterName  string `json:"tester_name,omitempty"`
	Location    string `json:"location,omitempty"`
	CompanyLogo string `json:"company_logo,omitempty"`
	OutputPath  string `json:"output_path,omitempty"`
}

// CaptivePortalReport contains all data for the report
type CaptivePortalReport struct {
	// Meta
	Config      CaptivePortalReportConfig `json:"config"`
	GeneratedAt time.Time                 `json:"generated_at"`

	// Target info
	TargetURL  string       `json:"target_url"`
	TargetName string       `json:"target_name"`
	PortalType string       `json:"portal_type"` // hotel, cafe, airport, etc.
	Analysis   FormAnalysis `json:"analysis"`

	// Attack summary
	TotalAttempts   int       `json:"total_attempts"`
	SuccessfulCodes []string  `json:"successful_codes,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Duration        string    `json:"duration"`

	// Bruteforce config used
	BruteConfig *BruteforceConfig `json:"brute_config,omitempty"`

	// Vulnerability assessment
	VulnerabilityLevel string    `json:"vulnerability_level"` // Critical, High, Medium, Low
	Findings           []Finding `json:"findings"`
	Recommendations    []string  `json:"recommendations"`
}

// Finding represents a security finding
type Finding struct {
	Title       string `json:"title"`
	Severity    string `json:"severity"` // Critical, High, Medium, Low, Info
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Remediation string `json:"remediation"`
}

// Color definitions
type rgbColor struct{ R, G, B int }

var (
	cpColorPrimary = rgbColor{41, 128, 185}
	cpColorDanger  = rgbColor{192, 57, 43}
	cpColorSuccess = rgbColor{39, 174, 96}
	cpColorWarning = rgbColor{243, 156, 18}
	cpColorDark    = rgbColor{44, 62, 80}
	cpColorLight   = rgbColor{236, 240, 241}
)

// GenerateCaptivePortalReport creates a PDF report
func GenerateCaptivePortalReport(report CaptivePortalReport) (string, error) {
	// Determine output path
	outputPath := report.Config.OutputPath
	if outputPath == "" {
		reportsDir := "/opt/heimdall/reports/captiveportal"
		if err := os.MkdirAll(reportsDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create reports directory: %v", err)
		}
		timestamp := time.Now().Format("20060102_150405")
		safeClientName := strings.ReplaceAll(report.Config.ClientName, " ", "_")
		outputPath = filepath.Join(reportsDir, fmt.Sprintf("rapport_captiveportal_%s_%s.pdf", safeClientName, timestamp))
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 20)

	// Title page
	pdf.AddPage()
	addCPTitlePage(pdf, report)

	// Executive summary
	pdf.AddPage()
	addCPExecutiveSummary(pdf, report)

	// Portal analysis
	pdf.AddPage()
	addCPPortalAnalysis(pdf, report)

	// Findings
	if len(report.Findings) > 0 {
		pdf.AddPage()
		addCPFindings(pdf, report)
	}

	// Recommendations
	pdf.AddPage()
	addCPRecommendations(pdf, report)

	// Save PDF
	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return "", fmt.Errorf("failed to save PDF: %v", err)
	}

	return outputPath, nil
}

func addCPTitlePage(pdf *gofpdf.Fpdf, report CaptivePortalReport) {
	pdf.Ln(40)

	// Title
	pdf.SetFont("Helvetica", "B", 28)
	pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
	pdf.CellFormat(0, 15, "RAPPORT D'AUDIT", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "B", 24)
	pdf.SetTextColor(cpColorPrimary.R, cpColorPrimary.G, cpColorPrimary.B)
	pdf.CellFormat(0, 12, "PORTAIL CAPTIF", "", 1, "C", false, 0, "")

	pdf.Ln(20)

	// Client info box
	pdf.SetFillColor(cpColorLight.R, cpColorLight.G, cpColorLight.B)
	pdf.Rect(40, pdf.GetY(), 130, 50, "F")

	pdf.SetXY(45, pdf.GetY()+5)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
	pdf.Cell(0, 8, "Client: "+report.Config.ClientName)

	pdf.SetXY(45, pdf.GetY()+10)
	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 7, "Cible: "+report.TargetName)

	pdf.SetXY(45, pdf.GetY()+8)
	pdf.Cell(0, 7, "URL: "+report.TargetURL)

	pdf.SetXY(45, pdf.GetY()+8)
	pdf.Cell(0, 7, fmt.Sprintf("Date: %s", report.GeneratedAt.Format("02/01/2006")))

	if report.Config.TesterName != "" {
		pdf.SetXY(45, pdf.GetY()+8)
		pdf.Cell(0, 7, "Testeur: "+report.Config.TesterName)
	}

	pdf.Ln(60)

	// Vulnerability level
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "Niveau de vulnérabilité:", "", 1, "C", false, 0, "")

	var levelColor rgbColor
	switch report.VulnerabilityLevel {
	case "Critical":
		levelColor = cpColorDanger
	case "High":
		levelColor = rgbColor{230, 126, 34}
	case "Medium":
		levelColor = cpColorWarning
	default:
		levelColor = cpColorSuccess
	}

	pdf.SetFont("Helvetica", "B", 24)
	pdf.SetTextColor(levelColor.R, levelColor.G, levelColor.B)
	pdf.CellFormat(0, 15, strings.ToUpper(report.VulnerabilityLevel), "", 1, "C", false, 0, "")
}

func addCPExecutiveSummary(pdf *gofpdf.Fpdf, report CaptivePortalReport) {
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
	pdf.CellFormat(0, 10, "Résumé Exécutif", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Summary box
	pdf.SetFillColor(cpColorLight.R, cpColorLight.G, cpColorLight.B)
	startY := pdf.GetY()
	pdf.Rect(15, startY, 180, 60, "F")

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetXY(20, startY+5)

	summary := fmt.Sprintf(`Cet audit de sécurité a ciblé le portail captif accessible à l'adresse %s. 
L'objectif était d'évaluer la robustesse des mécanismes d'authentification.

Durant cet audit:
• %d tentatives d'authentification ont été effectuées
• %d code(s)/mot(s) de passe valide(s) ont été découvert(s)
• Durée de l'audit: %s

Le niveau de risque global a été évalué comme %s.`,
		report.TargetURL,
		report.TotalAttempts,
		len(report.SuccessfulCodes),
		report.Duration,
		report.VulnerabilityLevel,
	)

	pdf.SetXY(20, startY+5)
	pdf.MultiCell(170, 6, summary, "", "L", false)

	pdf.SetY(startY + 65)

	// Statistics table
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 10, "Statistiques", "", 1, "L", false, 0, "")

	data := [][]string{
		{"Tentatives totales", fmt.Sprintf("%d", report.TotalAttempts)},
		{"Codes/MdP découverts", fmt.Sprintf("%d", len(report.SuccessfulCodes))},
		{"Durée de l'audit", report.Duration},
		{"Type de portail", report.PortalType},
	}

	pdf.SetFont("Helvetica", "", 11)
	for _, row := range data {
		pdf.SetFillColor(cpColorLight.R, cpColorLight.G, cpColorLight.B)
		pdf.CellFormat(90, 8, row[0], "1", 0, "L", true, 0, "")
		pdf.CellFormat(90, 8, row[1], "1", 1, "L", false, 0, "")
	}
}

func addCPPortalAnalysis(pdf *gofpdf.Fpdf, report CaptivePortalReport) {
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
	pdf.CellFormat(0, 10, "Analyse du Portail", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	analysis := report.Analysis
	pdf.SetFont("Helvetica", "", 11)

	// Form details
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Détails du formulaire:", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)

	formData := [][]string{
		{"Titre de la page", analysis.Title},
		{"Action du formulaire", analysis.FormAction},
		{"Méthode HTTP", analysis.FormMethod},
		{"Protection CSRF", boolToFrench(analysis.HasCSRF)},
		{"Captcha détecté", boolToFrench(analysis.HasCaptcha)},
	}

	for _, row := range formData {
		pdf.CellFormat(70, 7, row[0]+":", "", 0, "L", false, 0, "")
		pdf.CellFormat(110, 7, row[1], "", 1, "L", false, 0, "")
	}

	pdf.Ln(5)

	// Fields detected
	if len(analysis.Fields) > 0 {
		pdf.SetFont("Helvetica", "B", 12)
		pdf.CellFormat(0, 8, "Champs détectés:", "", 1, "L", false, 0, "")

		// Table header
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetFillColor(cpColorPrimary.R, cpColorPrimary.G, cpColorPrimary.B)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(50, 7, "Nom", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 7, "Type", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Requis", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Min", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Max", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Pattern", "1", 1, "C", true, 0, "")

		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)

		for i, field := range analysis.Fields {
			if field.Type == "hidden" {
				continue
			}
			fill := i%2 == 0
			if fill {
				pdf.SetFillColor(cpColorLight.R, cpColorLight.G, cpColorLight.B)
			}
			pdf.CellFormat(50, 6, field.Name, "1", 0, "L", fill, 0, "")
			pdf.CellFormat(30, 6, field.Type, "1", 0, "C", fill, 0, "")
			pdf.CellFormat(25, 6, boolToFrench(field.Required), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(25, 6, intToStr(field.MinLength), "1", 0, "C", fill, 0, "")
			pdf.CellFormat(25, 6, intToStr(field.MaxLength), "1", 0, "C", fill, 0, "")
			pattern := field.Pattern
			if len(pattern) > 10 {
				pattern = pattern[:10] + "..."
			}
			pdf.CellFormat(25, 6, pattern, "1", 1, "C", fill, 0, "")
		}
	}
}

func addCPFindings(pdf *gofpdf.Fpdf, report CaptivePortalReport) {
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
	pdf.CellFormat(0, 10, "Constats de Sécurité", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	for i, finding := range report.Findings {
		// Severity color
		var severityColor rgbColor
		switch finding.Severity {
		case "Critical":
			severityColor = cpColorDanger
		case "High":
			severityColor = rgbColor{230, 126, 34}
		case "Medium":
			severityColor = cpColorWarning
		case "Low":
			severityColor = cpColorSuccess
		default:
			severityColor = cpColorPrimary
		}

		// Finding header
		pdf.SetFillColor(severityColor.R, severityColor.G, severityColor.B)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 8, fmt.Sprintf("  %d. %s [%s]", i+1, finding.Title, finding.Severity), "1", 1, "L", true, 0, "")

		// Finding content
		pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
		pdf.SetFont("Helvetica", "", 10)

		pdf.SetFillColor(cpColorLight.R, cpColorLight.G, cpColorLight.B)
		startY := pdf.GetY()

		pdf.SetX(20)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Cell(30, 6, "Description:")
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(145, 5, finding.Description, "", "L", false)

		pdf.SetX(20)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Cell(30, 6, "Impact:")
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(145, 5, finding.Impact, "", "L", false)

		pdf.SetX(20)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Cell(30, 6, "Remédiation:")
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(145, 5, finding.Remediation, "", "L", false)

		endY := pdf.GetY()
		pdf.Rect(15, startY, 180, endY-startY, "D")

		pdf.Ln(5)
	}
}

func addCPRecommendations(pdf *gofpdf.Fpdf, report CaptivePortalReport) {
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
	pdf.CellFormat(0, 10, "Recommandations", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "", 11)
	for i, rec := range report.Recommendations {
		pdf.SetFillColor(cpColorPrimary.R, cpColorPrimary.G, cpColorPrimary.B)
		pdf.SetTextColor(255, 255, 255)
		pdf.CellFormat(8, 7, fmt.Sprintf("%d", i+1), "", 0, "C", true, 0, "")
		pdf.SetTextColor(cpColorDark.R, cpColorDark.G, cpColorDark.B)
		pdf.MultiCell(167, 7, "  "+rec, "", "L", false)
		pdf.Ln(2)
	}
}

// Helper functions
func boolToFrench(b bool) string {
	if b {
		return "Oui"
	}
	return "Non"
}

func intToStr(i int) string {
	if i == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", i)
}
