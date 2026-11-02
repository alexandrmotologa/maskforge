package discovery

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/maskforge/internal/faker"
)

// LeakViolation represents an audited finding where unmasked sensitive data was detected.
type LeakViolation struct {
	LineNumber    int64     `json:"line_number"`
	ViolationType string    `json:"violation_type"`
	Severity      RiskLevel `json:"severity"`
	Snippet       string    `json:"snippet"`
	Description   string    `json:"description"`
}

// VerificationReport encapsulates the result of a zero-leak post-masking audit.
type VerificationReport struct {
	TotalLinesScanned int64           `json:"total_lines_scanned"`
	Violations        []LeakViolation `json:"violations"`
	IsClean           bool            `json:"is_clean"`
	LeakScore         int             `json:"leak_score"` // 100 = perfectly clean, 0 = severe leaks
}

// VerifyOptions configures the strictness of the zero-leak verifier.
type VerifyOptions struct {
	AllowedDomains      []string
	AllowSyntheticCards bool
}

var (
	emailFinderRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@([a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})`)
	cardFinderRegex  = regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
	ssnFinderRegex   = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	secretKeyRegex   = regexp.MustCompile(`(?i)(BEGIN\s+PRIVATE\s+KEY|sk_live_[a-zA-Z0-9]{20,}|ghp_[a-zA-Z0-9]{20,})`)

	safeTestDomains = map[string]bool{
		"example.com":      true,
		"example.org":      true,
		"example.net":      true,
		"testmail.dev":     true,
		"staging.internal": true,
		"sandbox.local":    true,
	}
)

// VerifySQLDump scans an already-masked SQL stream and verifies zero PII leaks.
func VerifySQLDump(r io.Reader, opts VerifyOptions) (*VerificationReport, error) {
	allowedDomains := make(map[string]bool)
	for k, v := range safeTestDomains {
		allowedDomains[k] = v
	}
	for _, d := range opts.AllowedDomains {
		allowedDomains[strings.ToLower(strings.TrimSpace(d))] = true
	}

	report := &VerificationReport{
		Violations: make([]LeakViolation, 0),
		IsClean:    true,
		LeakScore:  100,
	}

	buf := bufio.NewReaderSize(r, 128*1024)
	var lineNum int64

	for {
		line, isPrefix, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		fullLine := line
		for isPrefix {
			var cont []byte
			cont, isPrefix, err = buf.ReadLine()
			if err != nil {
				return nil, err
			}
			fullLine = append(fullLine, cont...)
		}

		lineNum++
		lineStr := string(fullLine)
		trimmed := strings.TrimSpace(lineStr)

		// Skip comments and DDL definitions
		if strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(strings.ToUpper(trimmed), "CREATE TABLE") {
			continue
		}

		// 1. Audit Secrets & Private Keys
		if secretKeyRegex.MatchString(lineStr) {
			report.Violations = append(report.Violations, LeakViolation{
				LineNumber:    lineNum,
				ViolationType: "Exposed Secret / Key",
				Severity:      RiskCritical,
				Snippet:       truncateSnippet(lineStr, 60),
				Description:   "Potential exposed private key or live API secret token found in dump.",
			})
		}

		// 2. Audit Credit Card PANs
		if !opts.AllowSyntheticCards {
			cardMatches := cardFinderRegex.FindAllString(lineStr, -1)
			for _, cardCandidate := range cardMatches {
				if faker.IsValidLuhn(cardCandidate) {
					report.Violations = append(report.Violations, LeakViolation{
						LineNumber:    lineNum,
						ViolationType: "Luhn-Valid Credit Card",
						Severity:      RiskCritical,
						Snippet:       maskCardSnippet(cardCandidate),
						Description:   "16-digit credit card number passing Luhn checksum.",
					})
				}
			}
		}

		// 3. Audit Social Security Numbers
		ssnMatches := ssnFinderRegex.FindAllString(lineStr, -1)
		for _, ssnCandidate := range ssnMatches {
			if !strings.HasPrefix(ssnCandidate, "000") && !strings.HasPrefix(ssnCandidate, "666") {
				report.Violations = append(report.Violations, LeakViolation{
					LineNumber:    lineNum,
					ViolationType: "Social Security Number (SSN)",
					Severity:      RiskCritical,
					Snippet:       ssnCandidate[:3] + "-XX-XXXX",
					Description:   "Unmasked US Social Security Number detected in row data.",
				})
			}
		}

		// 4. Audit Real Emails
		emailMatches := emailFinderRegex.FindAllStringSubmatch(lineStr, -1)
		for _, emailMatch := range emailMatches {
			fullEmail := emailMatch[0]
			domain := strings.ToLower(emailMatch[1])

			if !allowedDomains[domain] {
				report.Violations = append(report.Violations, LeakViolation{
					LineNumber:    lineNum,
					ViolationType: "Production Email Address",
					Severity:      RiskHigh,
					Snippet:       fullEmail,
					Description:   "Email address from non-sandbox domain '" + domain + "'.",
				})
			}
		}
	}

	report.TotalLinesScanned = lineNum
	if len(report.Violations) > 0 {
		report.IsClean = false
		score := 100 - (len(report.Violations) * 15)
		if score < 0 {
			score = 0
		}
		report.LeakScore = score
	}

	return report, nil
}

func truncateSnippet(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func maskCardSnippet(card string) string {
	digits := strings.ReplaceAll(strings.ReplaceAll(card, "-", ""), " ", "")
	if len(digits) >= 8 {
		return digits[:4] + "********" + digits[len(digits)-4:]
	}
	return "4111********1111"
}
