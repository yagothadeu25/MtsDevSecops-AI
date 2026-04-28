package services

import (
	"fmt"
	
	"slices"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/server/logger"
	"pentagi/pkg/server/models"
	"pentagi/pkg/server/response"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jung-kurt/gofpdf/v2"
)

type ReportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

type reportData struct {
	Flow     models.Flow
	Tasks    []models.Task
	Subtasks []models.Subtask
	Reports  []models.Msglog
	Termlogs []models.Termlog
}

// GetFlowReportPDF generates a professional pentest PDF report for a flow
// @Summary Generate PDF report for a flow
// @Tags Flows,Reports
// @Produce application/pdf
// @Security BearerAuth
// @Param flowID path int true "flow id" minimum(0)
// @Success 200 {file} file "PDF report"
// @Failure 403 {object} response.errorResp "generating report not permitted"
// @Failure 404 {object} response.errorResp "flow not found"
// @Failure 500 {object} response.errorResp "internal error on generating report"
// @Router /flows/{flowID}/report/pdf [get]
func (s *ReportService) GetFlowReportPDF(c *gin.Context) {
	flowID, err := strconv.ParseUint(c.Param("flowID"), 10, 64)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error parsing flow id")
		response.Error(c, response.ErrFlowsInvalidRequest, err)
		return
	}

	uid := c.GetUint64("uid")
	privs := c.GetStringSlice("prm")
	var scope func(db *gorm.DB) *gorm.DB
	if slices.Contains(privs, "flows.admin") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ?", flowID)
		}
	} else if slices.Contains(privs, "flows.view") {
		scope = func(db *gorm.DB) *gorm.DB {
			return db.Where("id = ? AND user_id = ?", flowID, uid)
		}
	} else {
		response.Error(c, response.ErrNotPermitted, nil)
		return
	}

	data, err := s.collectReportData(flowID, scope)
	if err != nil {
		logger.FromContext(c).WithError(err).Errorf("error collecting report data")
		if gorm.IsRecordNotFoundError(err) {
			response.Error(c, response.ErrFlowsNotFound, err)
		} else {
			response.Error(c, response.ErrInternal, err)
		}
		return
	}

	pdf := s.generatePDF(data)

	filename := fmt.Sprintf("%d_%s_PenTest_%s-SerasaCyberShield.pdf",
		data.Flow.ID,
		sanitizeFilename(data.Flow.Title),
		time.Now().Format("01022006"),
	)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if err := pdf.Output(c.Writer); err != nil {
		logger.FromContext(c).WithError(err).Errorf("error writing PDF")
		return
	}
}

func (s *ReportService) collectReportData(flowID uint64, scope func(db *gorm.DB) *gorm.DB) (*reportData, error) {
	data := &reportData{}

	if err := s.db.Model(&data.Flow).Scopes(scope).Take(&data.Flow).Error; err != nil {
		return nil, err
	}

	s.db.Where("flow_id = ?", flowID).Order("created_at").Find(&data.Tasks)

	for _, t := range data.Tasks {
		var subs []models.Subtask
		s.db.Where("task_id = ?", t.ID).Order("created_at").Find(&subs)
		data.Subtasks = append(data.Subtasks, subs...)
	}

	s.db.Where("flow_id = ? AND type = ?", flowID, "report").Order("created_at").Find(&data.Reports)

	return data, nil
}

func (s *ReportService) generatePDF(data *reportData) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)

	s.addCoverPage(pdf, data)
	s.addTableOfContents(pdf, data)
	s.addExecutiveSummary(pdf, data)
	s.addScope(pdf, data)
	s.addFindings(pdf, data)
	s.addRecommendations(pdf, data)
	s.addMethodology(pdf)
	s.addDisclaimer(pdf)

	return pdf
}

// --- Cover Page ---

func (s *ReportService) addCoverPage(pdf *gofpdf.Fpdf, data *reportData) {
	pdf.AddPage()

	// Dark header bar
	pdf.SetFillColor(15, 23, 42) // slate-900
	pdf.Rect(0, 0, 210, 100, "F")

	// Brand
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 28)
	pdf.SetXY(20, 25)
	pdf.CellFormat(170, 12, "SERASA CYBER SHIELD AI", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 12)
	pdf.SetXY(20, 42)
	pdf.CellFormat(170, 8, encodeUTF8("Autonomous Penetration Testing Intelligence"), "", 1, "C", false, 0, "")

	// Accent line
	pdf.SetFillColor(59, 130, 246) // blue-500
	pdf.Rect(20, 58, 170, 2, "F")

	// Report type
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetXY(20, 68)
	pdf.CellFormat(170, 10, "PENETRATION TEST REPORT", "", 1, "C", false, 0, "")

	// Classification badge
	pdf.SetFillColor(220, 38, 38) // red-600
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetXY(70, 82)
	pdf.CellFormat(70, 8, "CONFIDENTIAL", "", 1, "C", true, 0, "")

	// Report details
	pdf.SetTextColor(30, 41, 59) // slate-800
	pdf.SetFont("Helvetica", "", 11)

	details := [][]string{
		{"Report Title:", data.Flow.Title},
		{"Report ID:", fmt.Sprintf("SCSA-%d-%s", data.Flow.ID, time.Now().Format("20060102"))},
		{"Date:", time.Now().Format("January 02, 2006")},
		{"Status:", strings.ToUpper(string(data.Flow.Status))},
		{"Model:", data.Flow.Model},
		{"Provider:", data.Flow.ModelProviderName},
	}

	y := 115.0
	for _, d := range details {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetXY(30, y)
		pdf.CellFormat(45, 7, d[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(120, 7, truncateStr(d[1], 80), "", 1, "L", false, 0, "")
		y += 8
	}

	// Footer
	pdf.SetFillColor(15, 23, 42)
	pdf.Rect(0, 270, 210, 27, "F")
	pdf.SetTextColor(148, 163, 184) // slate-400
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetXY(20, 275)
	pdf.CellFormat(170, 5, "This document contains confidential information. Unauthorized distribution is prohibited.", "", 1, "C", false, 0, "")
	pdf.SetXY(20, 281)
	pdf.CellFormat(170, 5, encodeUTF8("Generated by Serasa Cyber Shield AI - Autonomous Penetration Testing Platform"), "", 1, "C", false, 0, "")
}

// --- Table of Contents ---

func (s *ReportService) addTableOfContents(pdf *gofpdf.Fpdf, data *reportData) {
	pdf.AddPage()
	s.sectionHeader(pdf, "TABLE OF CONTENTS")

	items := []string{
		"1. Executive Summary",
		"2. Scope & Objectives",
		"3. Findings",
		"4. Recommendations",
		"5. Methodology",
		"6. Disclaimer",
	}

	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(30, 41, 59)
	for _, item := range items {
		pdf.SetX(30)
		pdf.CellFormat(150, 10, item, "", 1, "L", false, 0, "")
	}
}

// --- Executive Summary ---

func (s *ReportService) addExecutiveSummary(pdf *gofpdf.Fpdf, data *reportData) {
	pdf.AddPage()
	s.sectionHeader(pdf, "1. EXECUTIVE SUMMARY")

	// Stats box
	totalTasks := len(data.Tasks)
	totalSubtasks := len(data.Subtasks)
	totalReports := len(data.Reports)

	completedTasks := 0
	for _, t := range data.Tasks {
		if t.Status == "finished" {
			completedTasks++
		}
	}

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(30, 41, 59)

	// Summary stats table
	s.statsBox(pdf, "Tasks Executed", fmt.Sprintf("%d", totalTasks))
	s.statsBox(pdf, "Subtasks Completed", fmt.Sprintf("%d", totalSubtasks))
	s.statsBox(pdf, "Tasks Finished", fmt.Sprintf("%d/%d", completedTasks, totalTasks))
	s.statsBox(pdf, "Reports Generated", fmt.Sprintf("%d", totalReports))

	pdf.Ln(8)

	// Last report content as executive summary
	if len(data.Reports) > 0 {
		lastReport := data.Reports[len(data.Reports)-1]
		s.writeMarkdownContent(pdf, lastReport.Result)
	} else {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetTextColor(100, 116, 139)
		pdf.SetX(20)
		pdf.MultiCell(170, 6, "No report data available for this flow.", "", "L", false)
	}
}

// --- Scope ---

func (s *ReportService) addScope(pdf *gofpdf.Fpdf, data *reportData) {
	pdf.AddPage()
	s.sectionHeader(pdf, "2. SCOPE & OBJECTIVES")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(30, 41, 59)

	if len(data.Tasks) > 0 {
		for i, task := range data.Tasks {
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetX(20)
			pdf.CellFormat(170, 7, fmt.Sprintf("Task %d: %s", i+1, truncateStr(task.Title, 90)), "", 1, "L", false, 0, "")

			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(71, 85, 105)
			pdf.SetX(25)
			pdf.MultiCell(160, 5, fmt.Sprintf("Status: %s | Created: %s",
				strings.ToUpper(string(task.Status)),
				task.CreatedAt.Format("2006-01-02 15:04"),
			), "", "L", false)

			if task.Input != "" {
				pdf.SetX(25)
				pdf.MultiCell(160, 5, truncateStr(task.Input, 500), "", "L", false)
			}

			pdf.SetTextColor(30, 41, 59)
			pdf.Ln(4)
		}
	}

	// Subtasks summary
	if len(data.Subtasks) > 0 {
		pdf.Ln(4)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetX(20)
		pdf.CellFormat(170, 8, "Subtasks Breakdown", "", 1, "L", false, 0, "")

		// Table header
		pdf.SetFillColor(241, 245, 249) // slate-100
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetX(20)
		pdf.CellFormat(10, 7, "#", "1", 0, "C", true, 0, "")
		pdf.CellFormat(80, 7, "Title", "1", 0, "L", true, 0, "")
		pdf.CellFormat(20, 7, "Status", "1", 0, "C", true, 0, "")
		pdf.CellFormat(60, 7, "Result", "1", 1, "L", true, 0, "")

		pdf.SetFont("Helvetica", "", 8)
		for i, sub := range data.Subtasks {
			if pdf.GetY() > 260 {
				pdf.AddPage()
			}
			pdf.SetX(20)
			pdf.CellFormat(10, 6, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(80, 6, truncateStr(sub.Title, 50), "1", 0, "L", false, 0, "")
			pdf.CellFormat(20, 6, string(sub.Status), "1", 0, "C", false, 0, "")
			pdf.CellFormat(60, 6, truncateStr(sub.Result, 40), "1", 1, "L", false, 0, "")
		}
	}
}

// --- Findings ---

func (s *ReportService) addFindings(pdf *gofpdf.Fpdf, data *reportData) {
	pdf.AddPage()
	s.sectionHeader(pdf, "3. FINDINGS")

	if len(data.Reports) == 0 {
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetX(20)
		pdf.MultiCell(170, 6, "No findings were generated for this assessment.", "", "L", false)
		return
	}

	for i, report := range data.Reports {
		if pdf.GetY() > 240 {
			pdf.AddPage()
		}

		// Report header
		pdf.SetFillColor(241, 245, 249)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetX(20)
		pdf.CellFormat(170, 8, fmt.Sprintf("Report %d - %s",
			i+1,
			report.CreatedAt.Format("2006-01-02 15:04:05"),
		), "", 1, "L", true, 0, "")

		pdf.Ln(2)
		s.writeMarkdownContent(pdf, report.Result)
		pdf.Ln(6)
	}
}

// --- Recommendations ---

func (s *ReportService) addRecommendations(pdf *gofpdf.Fpdf, data *reportData) {
	pdf.AddPage()
	s.sectionHeader(pdf, "4. RECOMMENDATIONS")

	recommendations := []struct {
		priority string
		color    [3]int
		text     string
	}{
		{"CRITICAL", [3]int{220, 38, 38}, "Implement rate limiting on all authentication endpoints (5 attempts/min per IP)"},
		{"HIGH", [3]int{234, 88, 12}, "Standardize error responses to prevent information disclosure via status codes"},
		{"MEDIUM", [3]int{202, 138, 4}, "Add security headers: CSP, HSTS, X-Content-Type-Options, X-Frame-Options"},
		{"MEDIUM", [3]int{202, 138, 4}, "Ensure JSON parsing occurs after authentication validation"},
		{"LOW", [3]int{22, 163, 74}, "Implement monitoring and alerting for brute force patterns"},
	}

	for i, rec := range recommendations {
		if pdf.GetY() > 260 {
			pdf.AddPage()
		}

		// Priority badge
		pdf.SetFillColor(rec.color[0], rec.color[1], rec.color[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetX(20)
		pdf.CellFormat(22, 6, rec.priority, "", 0, "C", true, 0, "")

		// Recommendation text
		pdf.SetTextColor(30, 41, 59)
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(148, 6, fmt.Sprintf("  %d. %s", i+1, rec.text), "", 1, "L", false, 0, "")
		pdf.Ln(3)
	}
}

// --- Methodology ---

func (s *ReportService) addMethodology(pdf *gofpdf.Fpdf) {
	pdf.AddPage()
	s.sectionHeader(pdf, "5. METHODOLOGY")

	phases := []struct {
		name string
		desc string
	}{
		{"Reconnaissance", "Automated endpoint discovery, service fingerprinting, and attack surface mapping using integrated security tools."},
		{"Authentication & Authorization", "Testing authentication bypass, IDOR, token manipulation, and privilege escalation vectors."},
		{"Injection Testing", "SQL injection, NoSQL injection, command injection, XSS, and template injection across all input vectors."},
		{"Fuzzing & Resilience", "Parameter fuzzing, boundary testing, rate limiting validation, and error handling analysis."},
		{"Reporting", "AI-driven analysis and correlation of findings with CVSS v3.1 scoring and OWASP API Top 10 mapping."},
	}

	for i, phase := range phases {
		if pdf.GetY() > 260 {
			pdf.AddPage()
		}

		pdf.SetFillColor(59, 130, 246)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetX(20)
		pdf.CellFormat(8, 7, fmt.Sprintf("%d", i+1), "", 0, "C", true, 0, "")

		pdf.SetTextColor(30, 41, 59)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(162, 7, "  "+phase.name, "", 1, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(71, 85, 105)
		pdf.SetX(30)
		pdf.MultiCell(155, 5, phase.desc, "", "L", false)
		pdf.Ln(3)
	}

	pdf.Ln(6)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(30, 41, 59)
	pdf.SetX(20)
	pdf.CellFormat(170, 7, "Tools & Standards", "", 1, "L", false, 0, "")

	tools := []string{
		"Serasa Cyber Shield AI - Autonomous Penetration Testing Platform",
		"OWASP API Security Top 10 2023",
		"CVSS v3.1 Scoring Framework",
		"NIST Cybersecurity Framework",
		"nmap, hydra, sqlmap, ffuf, gobuster, dirb (containerized)",
	}

	pdf.SetFont("Helvetica", "", 9)
	for _, tool := range tools {
		pdf.SetX(25)
		pdf.CellFormat(165, 6, "- "+tool, "", 1, "L", false, 0, "")
	}
}

// --- Disclaimer ---

func (s *ReportService) addDisclaimer(pdf *gofpdf.Fpdf) {
	pdf.AddPage()
	s.sectionHeader(pdf, "6. DISCLAIMER")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(71, 85, 105)

	disclaimers := []string{
		"This penetration test report is provided for informational purposes only and is intended solely for the authorized recipient(s).",
		"The testing was performed within the agreed scope and timeframe. Findings represent the security posture at the time of testing and may not reflect current conditions.",
		"All testing was conducted in accordance with applicable laws and regulations. The tester assumes no liability for any damages resulting from the use of information contained in this report.",
		"This report should be treated as CONFIDENTIAL and should not be shared with unauthorized parties without prior written consent.",
		"The automated AI-driven testing methodology may not identify all vulnerabilities. Manual verification of critical findings is recommended.",
		"Remediation guidance is provided as general recommendations. Implementation should be validated by qualified security professionals.",
	}

	for _, d := range disclaimers {
		pdf.SetX(20)
		pdf.MultiCell(170, 5, "- "+d, "", "L", false)
		pdf.Ln(2)
	}
}

// --- Helpers ---

func (s *ReportService) sectionHeader(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFillColor(15, 23, 42)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetX(15)
	pdf.CellFormat(180, 10, "  "+title, "", 1, "L", true, 0, "")

	// Accent line
	pdf.SetFillColor(59, 130, 246)
	pdf.Rect(15, pdf.GetY(), 180, 1.5, "F")
	pdf.Ln(8)

	// Page footer
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(148, 163, 184)
		pdf.CellFormat(0, 10,
			fmt.Sprintf("Serasa Cyber Shield AI | CONFIDENTIAL | Page %d", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
}

func (s *ReportService) statsBox(pdf *gofpdf.Fpdf, label, value string) {
	x := pdf.GetX()
	if x < 25 {
		x = 20
	}

	pdf.SetFillColor(241, 245, 249)
	pdf.SetX(20)
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(85, 6, label, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(30, 41, 59)
	pdf.CellFormat(85, 6, value, "", 1, "R", false, 0, "")
}

func (s *ReportService) writeMarkdownContent(pdf *gofpdf.Fpdf, content string) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if pdf.GetY() > 265 {
			pdf.AddPage()
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			pdf.Ln(3)
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "### "):
			pdf.SetFont("Helvetica", "B", 10)
			pdf.SetTextColor(30, 41, 59)
			pdf.SetX(20)
			pdf.MultiCell(170, 6, strings.TrimPrefix(trimmed, "### "), "", "L", false)
		case strings.HasPrefix(trimmed, "## "):
			pdf.Ln(2)
			pdf.SetFont("Helvetica", "B", 11)
			pdf.SetTextColor(15, 23, 42)
			pdf.SetX(20)
			pdf.MultiCell(170, 7, strings.TrimPrefix(trimmed, "## "), "", "L", false)
		case strings.HasPrefix(trimmed, "# "):
			pdf.Ln(3)
			pdf.SetFont("Helvetica", "B", 13)
			pdf.SetTextColor(15, 23, 42)
			pdf.SetX(20)
			pdf.MultiCell(170, 8, strings.TrimPrefix(trimmed, "# "), "", "L", false)
		case strings.HasPrefix(trimmed, "- **") || strings.HasPrefix(trimmed, "* **"):
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(30, 41, 59)
			pdf.SetX(25)
			cleaned := strings.TrimLeft(trimmed, "-* ")
			cleaned = strings.ReplaceAll(cleaned, "**", "")
			pdf.MultiCell(160, 5, cleaned, "", "L", false)
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(71, 85, 105)
			pdf.SetX(25)
			cleaned := strings.TrimLeft(trimmed, "-* ")
			pdf.MultiCell(160, 5, cleaned, "", "L", false)
		case strings.HasPrefix(trimmed, "```"):
			// skip code fences
		case strings.HasPrefix(trimmed, "**") && strings.HasSuffix(trimmed, "**"):
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(30, 41, 59)
			pdf.SetX(20)
			cleaned := strings.Trim(trimmed, "*")
			pdf.MultiCell(170, 5, cleaned, "", "L", false)
		default:
			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(51, 65, 85)
			pdf.SetX(20)
			cleaned := strings.ReplaceAll(trimmed, "**", "")
			pdf.MultiCell(170, 5, cleaned, "", "L", false)
		}
	}
}

func sanitizeFilename(s string) string {
	replacer := strings.NewReplacer(
		" ", "_", "/", "_", "\\", "_",
		":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_",
		"|", "_",
	)
	result := replacer.Replace(s)
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func encodeUTF8(s string) string {
	return s
}
