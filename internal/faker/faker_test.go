package faker

import (
	"regexp"
	"strings"
	"testing"
)

func TestDeterministicRepeatability(t *testing.T) {
	salt := "test-secret-salt-12345"
	raw1 := "john.doe@enterprise.com"

	rng1 := NewDeterministicPRNG(salt, "email", raw1)
	out1 := GenerateEmail(rng1, raw1, false, "")

	rng2 := NewDeterministicPRNG(salt, "email", raw1)
	out2 := GenerateEmail(rng2, raw1, false, "")

	if out1 != out2 {
		t.Fatalf("expected deterministic output to match, got %q vs %q", out1, out2)
	}

	// Verify different raw input produces different output
	raw2 := "jane.smith@enterprise.com"
	rng3 := NewDeterministicPRNG(salt, "email", raw2)
	out3 := GenerateEmail(rng3, raw2, false, "")

	if out1 == out3 {
		t.Fatalf("expected different inputs to produce different outputs, got %q", out1)
	}
}

func TestReferentialIntegrityAcrossTables(t *testing.T) {
	reg := NewRegistry()
	salt := "shared-foreign-key-salt"
	userID := "usr_9984128"

	// Mask user_id in users table
	ctx1 := TransformContext{
		Salt:       salt,
		Table:      "users",
		Column:     "id",
		RawValue:   userID,
		RuleString: "faker.uuid",
	}
	maskedUser := reg.Transform(ctx1)

	// Mask customer_id in orders table referencing users.id
	ctx2 := TransformContext{
		Salt:       salt,
		Table:      "orders",
		Column:     "customer_id",
		RawValue:   userID,
		RuleString: "faker.uuid",
	}
	maskedOrderCustomer := reg.Transform(ctx2)

	if maskedUser != maskedOrderCustomer {
		t.Fatalf("referential integrity violated: users.id masked as %q but orders.customer_id masked as %q", maskedUser, maskedOrderCustomer)
	}
}

func TestCreditCardLuhnCompliance(t *testing.T) {
	salt := "cc-salt"
	for i := 0; i < 50; i++ {
		raw := string(rune('A'+i)) + "raw-card-sample"
		rng := NewDeterministicPRNG(salt, "credit_card", raw)
		card := GenerateCreditCard(rng, "any")

		if !IsValidLuhn(card) {
			t.Fatalf("generated credit card %q fails Luhn mod-10 check", card)
		}
	}
}

func TestEmailRFC5322(t *testing.T) {
	salt := "email-salt"
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

	for i := 0; i < 50; i++ {
		raw := "sample" + string(rune('a'+i)) + "@domain.org"
		rng := NewDeterministicPRNG(salt, "email", raw)
		masked := GenerateEmail(rng, raw, false, "")

		if !emailRegex.MatchString(masked) {
			t.Fatalf("generated email %q does not match RFC 5322 regex", masked)
		}
	}

	// Test preserve domain
	rng := NewDeterministicPRNG(salt, "email", "ceo@acme-corp.com")
	preserved := GenerateEmail(rng, "ceo@acme-corp.com", true, "")
	if !strings.HasSuffix(preserved, "@acme-corp.com") {
		t.Fatalf("expected domain to be preserved, got %q", preserved)
	}
}

func TestIBANValidity(t *testing.T) {
	salt := "iban-salt"
	for _, country := range []string{"RO", "DE", "FR", "GB"} {
		rng := NewDeterministicPRNG(salt, "iban", country)
		iban := GenerateIBAN(rng, country)

		if !strings.HasPrefix(iban, country) {
			t.Fatalf("expected IBAN to start with %s, got %s", country, iban)
		}
		if len(iban) < 15 {
			t.Fatalf("IBAN %s too short", iban)
		}
	}
}

func TestCNPValidity(t *testing.T) {
	salt := "cnp-salt"
	for i := 0; i < 20; i++ {
		rng := NewDeterministicPRNG(salt, "cnp", string(rune('0'+i)))
		cnp := GenerateCNP(rng)
		if len(cnp) != 13 {
			t.Fatalf("CNP %s length != 13", cnp)
		}
	}
}

func TestRegistryNullHandling(t *testing.T) {
	reg := NewRegistry()
	ctx := TransformContext{
		Salt:       "salt",
		Table:      "users",
		Column:     "email",
		RawValue:   `\N`,
		RuleString: "faker.email",
	}
	out := reg.Transform(ctx)
	if out != `\N` {
		t.Fatalf("expected \\N to be preserved, got %q", out)
	}

	ctx.RawValue = "NULL"
	out = reg.Transform(ctx)
	if out != "NULL" {
		t.Fatalf("expected NULL to be preserved, got %q", out)
	}
}

func TestParseRuleParameters(t *testing.T) {
	name, params := ParseRule("faker.email(preserve_domain=true, domain=test.org)")
	if name != "faker.email" {
		t.Fatalf("expected name faker.email, got %q", name)
	}
	if params["preserve_domain"] != "true" || params["domain"] != "test.org" {
		t.Fatalf("unexpected params: %+v", params)
	}

	name2, params2 := ParseRule("hash.constant('$2a$12$xyz')")
	if name2 != "hash.constant" {
		t.Fatalf("expected name hash.constant, got %q", name2)
	}
	if params2["value"] != "$2a$12$xyz" {
		t.Fatalf("expected value $2a$12$xyz, got %+v", params2)
	}
}
