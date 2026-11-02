package discovery

import (
	"strings"
	"testing"
)

func TestVerifySQLDumpClean(t *testing.T) {
	cleanSQL := `-- Clean masked dump
COPY public.users (id, email, phone) FROM stdin;
1	mary.smith@example.com	+40712345678
2	john.doe@testmail.dev	+12025550199
\.
`
	opts := VerifyOptions{}
	report, err := VerifySQLDump(strings.NewReader(cleanSQL), opts)
	if err != nil {
		t.Fatalf("verification error: %v", err)
	}

	if !report.IsClean {
		t.Fatalf("expected clean report, but got %d violations: %+v", len(report.Violations), report.Violations)
	}
	if report.LeakScore != 100 {
		t.Fatalf("expected leak score 100, got %d", report.LeakScore)
	}
}

func TestVerifySQLDumpDetectsLeaks(t *testing.T) {
	dirtySQL := `-- Production dump with leaks
COPY public.users (id, email, ssn) FROM stdin;
1	real.ceo@enterprise.com	123-45-6789
2	hacker@gmail.com	987-65-4321
\.
`
	opts := VerifyOptions{}
	report, err := VerifySQLDump(strings.NewReader(dirtySQL), opts)
	if err != nil {
		t.Fatalf("verification error: %v", err)
	}

	if report.IsClean {
		t.Fatalf("expected dirty report with violations, but got clean")
	}
	if len(report.Violations) < 2 {
		t.Fatalf("expected at least 2 violations, got %d", len(report.Violations))
	}
}
