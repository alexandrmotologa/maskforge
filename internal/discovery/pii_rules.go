package discovery

import (
	"regexp"
	"strings"

	"github.com/alexandrmotologa/maskforge/internal/faker"
)

// RiskLevel categorizes the sensitivity of detected PII.
type RiskLevel string

const (
	RiskCritical RiskLevel = "CRITICAL"
	RiskHigh     RiskLevel = "HIGH"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskLow      RiskLevel = "LOW"
)

// PIIRule represents a detection heuristic for sensitive data.
type PIIRule struct {
	Type            string
	NamePattern     *regexp.Regexp
	ValuePattern    *regexp.Regexp
	RecommendedRule string
	Risk            RiskLevel
	Compliance      []string // e.g. ["GDPR Art. 4", "PCI-DSS 3.4", "HIPAA § 164.514"]
	ValidateFunc    func(val string) bool
}

var defaultPIIRules = []PIIRule{
	{
		Type:            "Password / Secret",
		NamePattern:     regexp.MustCompile(`(?i).*(password|passwd|pwd|secret|auth_token|api_key|hash|salt|jwt|session_id).*`),
		RecommendedRule: "hash.constant",
		Risk:            RiskCritical,
		Compliance:      []string{"SOC 2", "ISO 27001", "PCI-DSS"},
	},
	{
		Type:            "Credit Card / PAN",
		NamePattern:     regexp.MustCompile(`(?i).*(credit_card|card_num|card_number|pan|cc_num|cvv|cvc).*`),
		ValuePattern:    regexp.MustCompile(`^\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}$`),
		RecommendedRule: "faker.credit_card",
		Risk:            RiskCritical,
		Compliance:      []string{"PCI-DSS Req 3.4", "GDPR Art. 4"},
		ValidateFunc:    faker.IsValidLuhn,
	},
	{
		Type:            "National ID / SSN",
		NamePattern:     regexp.MustCompile(`(?i).*(ssn|social_security|cnp|nin|pesel|tax_id|vat_id|fiscal_code|national_id|id_card).*`),
		ValuePattern:    regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$|^\d{13}$`),
		RecommendedRule: "faker.ssn",
		Risk:            RiskCritical,
		Compliance:      []string{"GDPR Art. 87", "HIPAA § 164.514"},
	},
	{
		Type:            "Email Address",
		NamePattern:     regexp.MustCompile(`(?i).*(email|mail|e_mail).*`),
		ValuePattern:    regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
		RecommendedRule: "faker.email",
		Risk:            RiskHigh,
		Compliance:      []string{"GDPR Art. 4", "HIPAA § 164.514"},
	},
	{
		Type:            "Phone Number",
		NamePattern:     regexp.MustCompile(`(?i).*(phone|mobile|cell|fax|telephone|tel_num).*`),
		ValuePattern:    regexp.MustCompile(`^\+?[0-9]{1,4}[-.\s]?\(?[0-9]{1,3}\)?[-.\s]?[0-9]{3,5}[-.\s]?[0-9]{4}$`),
		RecommendedRule: "faker.phone_number",
		Risk:            RiskHigh,
		Compliance:      []string{"GDPR Art. 4", "HIPAA § 164.514"},
	},
	{
		Type:            "Bank Account / IBAN",
		NamePattern:     regexp.MustCompile(`(?i).*(iban|account_num|bank_account|routing_number|bban).*`),
		ValuePattern:    regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}$`),
		RecommendedRule: "faker.iban",
		Risk:            RiskHigh,
		Compliance:      []string{"GDPR Art. 4", "PCI-DSS"},
	},
	{
		Type:            "First Name",
		NamePattern:     regexp.MustCompile(`(?i).*(first_name|firstname|fname|given_name|prenume).*`),
		RecommendedRule: "faker.first_name",
		Risk:            RiskMedium,
		Compliance:      []string{"GDPR Art. 4"},
	},
	{
		Type:            "Last Name",
		NamePattern:     regexp.MustCompile(`(?i).*(last_name|lastname|lname|surname|family_name).*`),
		RecommendedRule: "faker.last_name",
		Risk:            RiskMedium,
		Compliance:      []string{"GDPR Art. 4"},
	},
	{
		Type:            "Full Name",
		NamePattern:     regexp.MustCompile(`(?i).*(full_name|fullname|customer_name|contact_name|owner_name).*`),
		RecommendedRule: "faker.full_name",
		Risk:            RiskMedium,
		Compliance:      []string{"GDPR Art. 4"},
	},
	{
		Type:            "Physical Address",
		NamePattern:     regexp.MustCompile(`(?i).*(street|address|addr_line|home_address|billing_address|shipping_address).*`),
		RecommendedRule: "faker.street_address",
		Risk:            RiskMedium,
		Compliance:      []string{"GDPR Art. 4"},
	},
	{
		Type:            "City / Location",
		NamePattern:     regexp.MustCompile(`(?i).*(city|municipality|town|oras).*`),
		RecommendedRule: "faker.city",
		Risk:            RiskLow,
		Compliance:      []string{"GDPR Art. 4"},
	},
	{
		Type:            "Postal / ZIP Code",
		NamePattern:     regexp.MustCompile(`(?i).*(zip|zipcode|zip_code|postal_code|postcode).*`),
		ValuePattern:    regexp.MustCompile(`^\d{5}(-\d{4})?$|^\d{6}$`),
		RecommendedRule: "faker.zip_code",
		Risk:            RiskLow,
		Compliance:      []string{"GDPR Art. 4", "HIPAA § 164.514"},
	},
	{
		Type:            "IP Address",
		NamePattern:     regexp.MustCompile(`(?i).*(ip_address|client_ip|remote_ip|user_ip|ipv4|ipv6).*`),
		ValuePattern:    regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`),
		RecommendedRule: "faker.ipv4",
		Risk:            RiskLow,
		Compliance:      []string{"GDPR Recital 30"},
	},
	{
		Type:            "Date of Birth",
		NamePattern:     regexp.MustCompile(`(?i).*(dob|date_of_birth|birth_date|birthday).*`),
		RecommendedRule: "faker.date_of_birth",
		Risk:            RiskMedium,
		Compliance:      []string{"GDPR Art. 4", "HIPAA § 164.514"},
	},
}

// DetectColumn evaluates a column name and optional sample values to identify sensitive PII.
func DetectColumn(colName string, sampleValues []string) (match *PIIRule, confidence float64) {
	clean := strings.ToLower(strings.TrimSpace(colName))

	// 1. Column name heuristic
	for _, rule := range defaultPIIRules {
		if rule.NamePattern != nil && rule.NamePattern.MatchString(clean) {
			return &rule, 0.95
		}
	}

	// 2. Content pattern sniffing across sample values
	if len(sampleValues) > 0 {
		for _, rule := range defaultPIIRules {
			if rule.ValuePattern != nil || rule.ValidateFunc != nil {
				validCount := 0
				testedCount := 0
				for _, val := range sampleValues {
					val = strings.Trim(strings.TrimSpace(val), "'\"")
					if val == "" || val == `\N` || val == "NULL" {
						continue
					}
					testedCount++
					matched := false
					if rule.ValuePattern != nil && rule.ValuePattern.MatchString(val) {
						matched = true
					}
					if matched && rule.ValidateFunc != nil {
						matched = rule.ValidateFunc(val)
					}
					if matched {
						validCount++
					}
				}

				if testedCount > 0 && float64(validCount)/float64(testedCount) >= 0.7 {
					return &rule, 0.85
				}
			}
		}
	}

	return nil, 0.0
}
