package eviltwin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// EvilTwinReport represents an Evil Twin attack report
type EvilTwinReport struct {
	ClientName  string               `json:"client_name"`
	TesterName  string               `json:"tester_name"`
	Objective   string               `json:"objective"`
	TargetSSID  string               `json:"target_ssid"`
	FakeSSID    string               `json:"fake_ssid"`
	StartedAt   time.Time            `json:"started_at"`
	EndedAt     time.Time            `json:"ended_at"`
	Credentials []CapturedCredential `json:"credentials"`
	DeauthSent  int                  `json:"deauth_sent"`
}

// GenerateEvilTwinReport generates a PDF report for the Evil Twin attack
func GenerateEvilTwinReport(report EvilTwinReport) (string, error) {
	reportsDir := "/opt/heimdall/reports/eviltwin"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("impossible de créer le répertoire: %v", err)
	}

	filename := fmt.Sprintf("eviltwin_report_%s_%s.pdf",
		report.ClientName,
		report.EndedAt.Format("20060102_150405"))
	filepath := filepath.Join(reportsDir, filename)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)

	// Title page
	addETTitlePage(pdf, report)

	// Executive Summary
	addETExecutiveSummary(pdf, report)

	// Attack Details
	addETAttackDetails(pdf, report)

	// Captured Credentials
	addETCredentials(pdf, report)

	// Security Recommendations
	addETRecommendations(pdf, report)

	// Awareness Section
	addETAwarenessSection(pdf, report)

	if err := pdf.OutputFileAndClose(filepath); err != nil {
		return "", fmt.Errorf("erreur de génération du PDF: %v", err)
	}

	return filepath, nil
}

func addETTitlePage(pdf *gofpdf.Fpdf, report EvilTwinReport) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 28)

	// Title
	pdf.SetY(80)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 15, "RAPPORT D'ATTAQUE EVIL TWIN", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 16)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 10, "Test de sensibilisation WiFi", "", 1, "C", false, 0, "")

	// Warning banner
	pdf.SetY(120)
	pdf.SetFillColor(255, 200, 200)
	pdf.SetTextColor(150, 50, 50)
	pdf.Rect(20, 120, 170, 25, "F")
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetXY(25, 125)
	pdf.MultiCell(160, 6, "ATTENTION: Ce rapport contient des informations sensibles.\nA manipuler avec précaution et à ne pas diffuser.", "", "C", false)

	// Info box
	pdf.SetY(160)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, fmt.Sprintf("Client: %s", report.ClientName), "", 1, "C", false, 0, "")
	if report.TesterName != "" {
		pdf.CellFormat(0, 8, fmt.Sprintf("Auditeur: %s", report.TesterName), "", 1, "C", false, 0, "")
	}
	pdf.CellFormat(0, 8, fmt.Sprintf("Date: %s", report.EndedAt.Format("02/01/2006")), "", 1, "C", false, 0, "")

	// Objective
	objectiveLabels := map[string]string{
		"awareness": "Sensibilisation des utilisateurs",
		"audit":     "Audit de sécurité WiFi",
		"pentest":   "Test de pénétration",
	}
	objLabel := objectiveLabels[report.Objective]
	if objLabel == "" {
		objLabel = "Évaluation de sécurité"
	}
	pdf.SetY(200)
	pdf.SetFont("Helvetica", "I", 11)
	pdf.CellFormat(0, 8, fmt.Sprintf("Objectif: %s", objLabel), "", 1, "C", false, 0, "")

	// Footer
	pdf.SetY(270)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 5, "Généré par Heimdall Security Platform", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, "https://heimdall.local", "", 1, "C", false, 0, "")
}

func addETExecutiveSummary(pdf *gofpdf.Fpdf, report EvilTwinReport) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 10, "Résumé Exécutif", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(60, 60, 60)

	duration := report.EndedAt.Sub(report.StartedAt)
	credsCount := len(report.Credentials)

	// Summary text
	summaryText := fmt.Sprintf(
		"Dans le cadre de l'évaluation de la sensibilisation aux risques WiFi pour %s, "+
			"une attaque de type 'Evil Twin' a été réalisée le %s.\n\n"+
			"Cette technique consiste à créer un faux point d'accès WiFi imitant un réseau légitime "+
			"afin de capturer les identifiants des utilisateurs qui s'y connectent.\n\n"+
			"RÉSULTATS:\n"+
			"• Durée de l'attaque: %s\n"+
			"• Réseau ciblé: %s\n"+
			"• Faux réseau créé: %s\n"+
			"• Paquets de déauthentification envoyés: %d\n"+
			"• Identifiants capturés: %d\n\n",
		report.ClientName,
		report.StartedAt.Format("02/01/2006 à 15:04"),
		formatDuration(duration),
		report.TargetSSID,
		report.FakeSSID,
		report.DeauthSent,
		credsCount,
	)

	pdf.MultiCell(0, 6, summaryText, "", "L", false)

	// Risk level
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "B", 12)
	
	var riskLevel, riskColor string
	if credsCount == 0 {
		riskLevel = "FAIBLE"
		riskColor = "G"
		pdf.SetTextColor(50, 150, 50)
	} else if credsCount < 5 {
		riskLevel = "MOYEN"
		riskColor = "O"
		pdf.SetTextColor(200, 150, 50)
	} else {
		riskLevel = "ÉLEVÉ"
		riskColor = "R"
		pdf.SetTextColor(200, 50, 50)
	}

	pdf.CellFormat(0, 8, fmt.Sprintf("Niveau de risque évalué: %s", riskLevel), "", 1, "L", false, 0, "")
	
	_ = riskColor // Used for potential visual indicator
}

func addETAttackDetails(pdf *gofpdf.Fpdf, report EvilTwinReport) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 10, "Détails de l'Attaque", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(60, 60, 60)

	// Attack methodology
	methodText := `L'attaque Evil Twin (Jumeau Maléfique) est une technique de piratage WiFi consistant à:

1. RECONNAISSANCE
   - Identification du réseau WiFi cible et de ses caractéristiques
   - Analyse du portail captif (si présent)
   - Détermination du canal de diffusion

2. PRÉPARATION
   - Clonage du portail captif du réseau légitime
   - Configuration d'un point d'accès identique (même SSID, même canal)
   - Mise en place d'un serveur DHCP et DNS malveillant

3. DÉAUTHENTIFICATION
   - Envoi de paquets de déconnexion aux clients légitimes
   - Forçage de la reconnexion sur le faux réseau

4. CAPTURE
   - Les victimes se connectent au faux réseau
   - Présentation du portail captif cloné
   - Capture des identifiants saisis

5. ANALYSE
   - Extraction et analyse des données capturées
   - Génération du rapport de sensibilisation
`

	pdf.MultiCell(0, 6, methodText, "", "L", false)

	// Timeline
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, "Chronologie", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)

	pdf.CellFormat(0, 6, fmt.Sprintf("• Début de l'attaque: %s", report.StartedAt.Format("02/01/2006 15:04:05")), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("• Fin de l'attaque: %s", report.EndedAt.Format("02/01/2006 15:04:05")), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("• Durée totale: %s", formatDuration(report.EndedAt.Sub(report.StartedAt))), "", 1, "L", false, 0, "")
}

func addETCredentials(pdf *gofpdf.Fpdf, report EvilTwinReport) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 10, "Identifiants Capturés", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	if len(report.Credentials) == 0 {
		pdf.SetFont("Helvetica", "I", 12)
		pdf.SetTextColor(100, 150, 100)
		pdf.CellFormat(0, 10, "✓ Aucun identifiant n'a été capturé lors de ce test.", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		pdf.SetFont("Helvetica", "", 11)
		pdf.SetTextColor(60, 60, 60)
		pdf.MultiCell(0, 6, "Cela peut indiquer:\n"+
			"• Une bonne sensibilisation des utilisateurs\n"+
			"• Un faible trafic pendant la période de test\n"+
			"• Des mesures de sécurité efficaces (802.1X, WPA3)\n"+
			"• Détection de l'attaque par l'infrastructure", "", "L", false)
		return
	}

	// Warning
	pdf.SetFillColor(255, 240, 240)
	pdf.SetTextColor(150, 50, 50)
	pdf.Rect(20, pdf.GetY(), 170, 15, "F")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(25, pdf.GetY()+3)
	pdf.MultiCell(160, 5, "ATTENTION: Ces informations sont confidentielles.\nLes mots de passe doivent être réinitialisés.", "", "C", false)
	pdf.Ln(10)

	// Table header
	pdf.SetFillColor(60, 60, 60)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(35, 8, "Heure", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "IP Source", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Login", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 8, "Mot de passe", "1", 1, "C", true, 0, "")

	// Table rows
	pdf.SetFillColor(245, 245, 245)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Helvetica", "", 9)

	for i, cred := range report.Credentials {
		fill := i%2 == 0
		pdf.CellFormat(35, 7, cred.Timestamp.Format("15:04:05"), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(35, 7, truncateString(cred.SourceIP, 15), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(50, 7, truncateString(cred.Login, 20), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(50, 7, maskPassword(cred.Password), "1", 1, "C", fill, 0, "")
	}

	// Statistics
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 8, fmt.Sprintf("Total: %d identifiants capturés", len(report.Credentials)), "", 1, "L", false, 0, "")
}

func addETRecommendations(pdf *gofpdf.Fpdf, report EvilTwinReport) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 10, "Recommandations de Sécurité", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(60, 60, 60)

	recommendations := []struct {
		title string
		desc  string
		priority string
	}{
		{
			title: "1. Implémenter WPA3 ou 802.1X",
			desc:  "Migrer vers WPA3-Enterprise ou utiliser l'authentification 802.1X (RADIUS) qui rend les attaques Evil Twin inefficaces car le client vérifie l'authenticité du point d'accès.",
			priority: "HAUTE",
		},
		{
			title: "2. Utiliser des certificats côté client",
			desc:  "Déployer des certificats sur les appareils des utilisateurs pour authentifier le réseau WiFi. Cela empêche la connexion à des réseaux non autorisés.",
			priority: "HAUTE",
		},
		{
			title: "3. Désactiver la connexion automatique",
			desc:  "Configurer les appareils pour ne pas se connecter automatiquement aux réseaux connus. L'utilisateur doit valider manuellement chaque connexion.",
			priority: "MOYENNE",
		},
		{
			title: "4. Sensibilisation des utilisateurs",
			desc:  "Former les utilisateurs à reconnaître les signes d'un réseau malveillant: déconnexions répétées, portail captif inattendu, certificat invalide.",
			priority: "HAUTE",
		},
		{
			title: "5. Surveillance réseau",
			desc:  "Déployer un système de détection d'intrusion sans fil (WIDS) capable de détecter les points d'accès pirates et les attaques de déauthentification.",
			priority: "MOYENNE",
		},
		{
			title: "6. Utiliser un VPN",
			desc:  "Imposer l'utilisation d'un VPN d'entreprise qui chiffre tout le trafic, même si l'utilisateur se connecte à un réseau compromis.",
			priority: "MOYENNE",
		},
		{
			title: "7. Réinitialiser les mots de passe compromis",
			desc:  "Les utilisateurs dont les identifiants ont été capturés lors de ce test doivent immédiatement changer leurs mots de passe.",
			priority: "CRITIQUE",
		},
	}

	for _, rec := range recommendations {
		// Priority badge
		pdf.SetFont("Helvetica", "B", 10)
		switch rec.priority {
		case "CRITIQUE":
			pdf.SetTextColor(200, 0, 0)
		case "HAUTE":
			pdf.SetTextColor(200, 100, 0)
		case "MOYENNE":
			pdf.SetTextColor(150, 150, 0)
		default:
			pdf.SetTextColor(100, 100, 100)
		}
		pdf.CellFormat(20, 6, rec.priority, "", 0, "L", false, 0, "")

		// Title
		pdf.SetTextColor(50, 50, 50)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 6, rec.title, "", 1, "L", false, 0, "")

		// Description
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(80, 80, 80)
		pdf.MultiCell(0, 5, rec.desc, "", "L", false)
		pdf.Ln(3)
	}
}

func addETAwarenessSection(pdf *gofpdf.Fpdf, report EvilTwinReport) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 10, "Matériel de Sensibilisation", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(60, 60, 60)

	awarenessText := `Ce document peut être utilisé comme support de formation pour sensibiliser 
les utilisateurs aux risques des réseaux WiFi publics et des attaques Evil Twin.

POINTS CLÉS À COMMUNIQUER:

🔐 NE JAMAIS se connecter à un réseau WiFi inconnu
   Même si le nom semble familier, vérifiez son authenticité.

⚠️ MÉFIEZ-VOUS des déconnexions répétées
   Si vous êtes déconnecté plusieurs fois de suite, c'est peut-être 
   le signe d'une attaque de déauthentification.

🔒 VÉRIFIEZ toujours le certificat
   Si votre navigateur ou appareil signale un problème de certificat, 
   ne continuez pas.

📱 NE SAISISSEZ JAMAIS d'identifiants sensibles sur un réseau public
   Attendez d'être sur un réseau de confiance ou utilisez votre 4G/5G.

🛡️ UTILISEZ un VPN
   Un VPN chiffre votre trafic et vous protège même sur un réseau compromis.

❌ NE FAITES PAS confiance aux portails captifs
   Un attaquant peut cloner n'importe quel portail d'authentification.


CE QU'IL S'EST PASSÉ LORS DE CE TEST:

Un point d'accès WiFi malveillant a été créé avec un nom similaire au 
réseau légitime. Les utilisateurs qui s'y sont connectés ont été 
redirigés vers un faux portail captif où leurs identifiants ont été capturés.

Cette démonstration illustre la facilité avec laquelle un attaquant peut 
voler des identifiants sur un réseau WiFi non sécurisé.
`

	pdf.MultiCell(0, 6, awarenessText, "", "L", false)

	// Signature block
	pdf.Ln(20)
	pdf.SetFont("Helvetica", "I", 10)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 6, "Document généré automatiquement par Heimdall Security Platform", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Date de génération: %s", time.Now().Format("02/01/2006 15:04")), "", 1, "C", false, 0, "")
}

// Helper functions

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func maskPassword(password string) string {
	if len(password) <= 2 {
		return "***"
	}
	// Show first and last character, mask the rest
	return password[:1] + strings.Repeat("*", len(password)-2) + password[len(password)-1:]
}
