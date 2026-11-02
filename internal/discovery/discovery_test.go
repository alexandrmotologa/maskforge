package discovery

import (
	"strings"
	"testing"
)

func TestDetectColumnHeuristics(t *testing.T) {
	tests := []struct {
		colName  string
		expected string
		risk     RiskLevel
	}{
		{"email", "faker.email", RiskHigh},
		{"contact_email", "faker.email", RiskHigh},
		{"user_phone", "faker.phone_number", RiskHigh},
		{"first_name", "faker.first_name", RiskMedium},
		{"last_name", "faker.last_name", RiskMedium},
		{"credit_card_number", "faker.credit_card", RiskCritical},
		{"ssn", "faker.ssn", RiskCritical},
		{"bank_iban", "faker.iban", RiskHigh},
		{"password_hash", "hash.constant", RiskCritical},
		{"billing_address", "faker.street_address", RiskMedium},
		{"client_ip", "faker.ipv4", RiskLow},
	}

	for _, tt := range tests {
		rule, _ := DetectColumn(tt.colName, nil)
		if rule == nil {
			t.Fatalf("expected detection for %s, got nil", tt.colName)
		}
		if rule.RecommendedRule != tt.expected {
			t.Fatalf("for %s expected rule %s, got %s", tt.colName, tt.expected, rule.RecommendedRule)
		}
		if rule.Risk != tt.risk {
			t.Fatalf("for %s expected risk %s, got %s", tt.colName, tt.risk, rule.Risk)
		}
	}
}

func TestContentSniffingAmbiguousColumn(t *testing.T) {
	samples := []string{
		"alice@company.com",
		"bob@test.org",
		"carol.danvers@domain.io",
	}

	rule, conf := DetectColumn("user_data_field", samples)
	if rule == nil {
		t.Fatalf("expected content sniffing to match email pattern, got nil")
	}
	if rule.RecommendedRule != "faker.email" {
		t.Fatalf("expected faker.email, got %s", rule.RecommendedRule)
	}
	if conf < 0.8 {
		t.Fatalf("expected confidence >= 0.8, got %f", conf)
	}
}

func TestScanReaderAndConfigGeneration(t *testing.T) {
	dumpSQL := `
CREATE TABLE users (
    id integer NOT NULL,
    email character varying(255),
    full_name character varying(255),
    card_number character varying(32)
);

COPY users (id, email, full_name, card_number) FROM stdin;
1	mark@acme.com	Mark Zuckerberg	4111111111111111
2	satya@microsoft.com	Satya Nadella	5500000000000004
\.
`

	scanner := NewSchemaScanner("test-scan-salt")
	report, err := scanner.ScanReader(strings.NewReader(dumpSQL), 5)
	if err != nil {
		t.Fatalf("scan reader failed: %v", err)
	}

	if report.TotalTables != 1 {
		t.Fatalf("expected 1 table, got %d", report.TotalTables)
	}
	if report.TotalColumns != 4 {
		t.Fatalf("expected 4 columns, got %d", report.TotalColumns)
	}
	if report.PIIColumns != 3 {
		t.Fatalf("expected 3 PII columns (email, full_name, card_number), got %d", report.PIIColumns)
	}
	if report.EstimatedRisk != RiskCritical {
		t.Fatalf("expected CRITICAL risk due to card_number, got %s", report.EstimatedRisk)
	}

	cfg := GenerateConfigFromReport(report)
	if cfg == nil {
		t.Fatalf("expected generated config, got nil")
	}

	usersCfg, ok := cfg.Tables["users"]
	if !ok {
		t.Fatalf("expected users table in generated config")
	}
	if usersCfg.Columns["email"] != "faker.email" {
		t.Fatalf("expected faker.email for users.email, got %s", usersCfg.Columns["email"])
	}
	if usersCfg.Columns["card_number"] != "faker.credit_card" {
		t.Fatalf("expected faker.credit_card for users.card_number, got %s", usersCfg.Columns["card_number"])
	}
}
