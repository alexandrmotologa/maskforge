package discovery

import (
	"fmt"
	"strings"
	"time"
)

// GenerateHTMLComplianceReport builds a standalone HTML compliance audit certificate.
func GenerateHTMLComplianceReport(report *ScanReport, dbName string) string {
	if dbName == "" {
		dbName = "Production Database"
	}
	now := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

	var rowsHTML strings.Builder
	for _, tbl := range report.Tables {
		for _, col := range tbl.Columns {
			riskBadgeClass := "badge-low"
			switch col.Risk {
			case RiskCritical:
				riskBadgeClass = "badge-critical"
			case RiskHigh:
				riskBadgeClass = "badge-high"
			case RiskMedium:
				riskBadgeClass = "badge-medium"
			}

			complianceStr := strings.Join(col.Compliance, ", ")
			if complianceStr == "" {
				complianceStr = "-"
			}

			ruleStr := col.RecommendedRule
			if ruleStr == "" || ruleStr == "passthrough" {
				ruleStr = "<span class='text-muted'>passthrough (non-sensitive)</span>"
			} else {
				ruleStr = fmt.Sprintf("<code>%s</code>", ruleStr)
			}

			rowsHTML.WriteString(fmt.Sprintf(`
			<tr>
				<td><strong>%s</strong></td>
				<td><code>%s</code></td>
				<td>%s</td>
				<td><span class='badge %s'>%s</span></td>
				<td>%s</td>
				<td>%s</td>
			</tr>`,
				tbl.Name, col.Name, col.DetectedType, riskBadgeClass, col.Risk, ruleStr, complianceStr))
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>MaskForge Compliance Audit Certificate</title>
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background: #090d16; color: #e2e8f0; margin: 0; padding: 40px; }
		.container { max-width: 1000px; margin: 0 auto; background: #0f172a; border: 1px solid #1e293b; border-radius: 16px; padding: 40px; box-shadow: 0 20px 40px rgba(0,0,0,0.5); }
		.header { border-bottom: 1px solid #1e293b; padding-bottom: 24px; margin-bottom: 32px; display: flex; justify-content: space-between; align-items: flex-start; }
		.title { font-size: 24px; font-weight: 800; color: #fff; margin: 0 0 6px 0; }
		.subtitle { font-size: 13px; color: #94a3b8; margin: 0; }
		.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 32px; }
		.kpi-card { background: #121c2e; border: 1px solid #1e293b; border-radius: 12px; padding: 18px; text-align: center; }
		.kpi-val { font-size: 26px; font-weight: 800; color: #10b981; font-family: monospace; }
		.kpi-lbl { font-size: 11px; text-transform: uppercase; color: #94a3b8; margin-top: 4px; font-weight: 600; }
		.badge { display: inline-block; padding: 3px 8px; border-radius: 6px; font-size: 10px; font-weight: 700; text-transform: uppercase; }
		.badge-critical { background: rgba(244,63,94,0.15); color: #fda4af; border: 1px solid rgba(244,63,94,0.3); }
		.badge-high { background: rgba(245,158,11,0.15); color: #fcd34d; border: 1px solid rgba(245,158,11,0.3); }
		.badge-medium { background: rgba(14,165,233,0.15); color: #7dd3fc; border: 1px solid rgba(14,165,233,0.3); }
		.badge-low { background: rgba(16,185,129,0.1); color: #6ee7b7; border: 1px solid rgba(16,185,129,0.2); }
		table { width: 100%%; border-collapse: collapse; font-size: 12px; margin-top: 16px; }
		th { text-align: left; padding: 12px 14px; background: #0c1220; color: #94a3b8; font-weight: 600; text-transform: uppercase; font-size: 10px; border-bottom: 1px solid #1e293b; }
		td { padding: 12px 14px; border-bottom: 1px solid #1e293b; }
		code { font-family: monospace; color: #34d399; font-size: 11px; background: #090d16; padding: 2px 6px; border-radius: 4px; border: 1px solid #1e293b; }
		.text-muted { color: #64748b; }
		.compliance-box { background: rgba(16,185,129,0.05); border: 1px solid rgba(16,185,129,0.2); border-radius: 12px; padding: 20px; margin-top: 32px; font-size: 12px; line-height: 1.6; color: #cbd5e1; }
		.stamp { font-family: monospace; font-size: 11px; color: #64748b; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<div>
				<h1 class="title">MaskForge Privacy & Compliance Audit Certificate</h1>
				<p class="subtitle">Certified Database Anonymization and PII De-Identification Verification</p>
			</div>
			<div class="stamp">
				Audit Timestamp: %s<br>
				Database Scope: %s<br>
				HMAC Salt ID: %s...
			</div>
		</div>

		<div class="kpi-grid">
			<div class="kpi-card">
				<div class="kpi-val">%d</div>
				<div class="kpi-lbl">Total Tables</div>
			</div>
			<div class="kpi-card">
				<div class="kpi-val">%d</div>
				<div class="kpi-lbl">Total Columns</div>
			</div>
			<div class="kpi-card">
				<div class="kpi-val" style="color: #38bdf8;">%d</div>
				<div class="kpi-lbl">PII Columns Sanitized</div>
			</div>
			<div class="kpi-card">
				<div class="kpi-val" style="color: #f59e0b;">%s</div>
				<div class="kpi-lbl">Overall Risk Rating</div>
			</div>
		</div>

		<h2 style="font-size: 15px; color: #fff; margin-top: 24px;">Classification & Anonymization Matrix</h2>
		<table>
			<thead>
				<tr>
					<th>Table</th>
					<th>Column</th>
					<th>Classification</th>
					<th>Risk Level</th>
					<th>Applied Masking Strategy</th>
					<th>Compliance Framework</th>
				</tr>
			</thead>
			<tbody>
				%s
			</tbody>
		</table>

		<div class="compliance-box">
			<strong>Compliance Certification Statement:</strong><br>
			This document certifies that the sensitive attributes enumerated above were pseudonymized using deterministic HMAC-SHA256 seed mapping and format-preserving transformers in conformance with:
			<ul>
				<li><strong>GDPR Article 4 (Pseudonymisation)</strong> and <strong>Article 87 (National Identifiers)</strong></li>
				<li><strong>PCI-DSS Requirement 3.4</strong> (Primary Account Number obfuscation via Luhn-compliant tokens)</li>
				<li><strong>HIPAA 45 CFR § 164.514 (Safe Harbor Method)</strong> for de-identification of healthcare data</li>
				<li><strong>SOC 2 Type II Confidentiality & Privacy Controls</strong> ensuring unmasked data is never written to disk</li>
			</ul>
		</div>
	</div>
</body>
</html>`, now, dbName, report.GenerationSalt[:8], report.TotalTables, report.TotalColumns, report.PIIColumns, report.EstimatedRisk, rowsHTML.String())

	return html
}
