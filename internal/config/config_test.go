package config

import (
	"testing"
)

const sampleValidYAML = `
version: "1.0"
salt: "test-static-salt"
global:
  preserve_nulls: true
  ignore_columns: ["id", "created_at"]
tables:
  users:
    columns:
      email: "faker.email"
      first_name: "faker.first_name"
      last_name: "faker.last_name"
      phone: "faker.phone_number"
      password_hash: "hash.constant($2a$12$stagingHash)"
    ignore_columns: ["tenant_id"]
  orders:
    columns:
      customer_email: "faker.email"
      credit_card: "faker.credit_card(brand=visa)"
`

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadFromBytes([]byte(sampleValidYAML))
	if err != nil {
		t.Fatalf("failed to parse valid YAML: %v", err)
	}

	if cfg.Salt != "test-static-salt" {
		t.Fatalf("expected salt test-static-salt, got %s", cfg.Salt)
	}

	// Test rule resolution on users
	rule, mask := cfg.GetColumnRule("users", "email")
	if !mask || rule != "faker.email" {
		t.Fatalf("expected users.email to mask with faker.email, got mask=%v rule=%s", mask, rule)
	}

	// Test qualified table name "public.users"
	rule2, mask2 := cfg.GetColumnRule("public.users", "email")
	if !mask2 || rule2 != "faker.email" {
		t.Fatalf("expected public.users.email to resolve, got mask=%v rule=%s", mask2, rule2)
	}

	// Test global ignored column
	ruleIgn, maskIgn := cfg.GetColumnRule("users", "created_at")
	if maskIgn || ruleIgn != "passthrough" {
		t.Fatalf("expected created_at to be ignored, got mask=%v rule=%s", maskIgn, ruleIgn)
	}

	// Test table-level ignored column
	ruleIgn2, maskIgn2 := cfg.GetColumnRule("users", "tenant_id")
	if maskIgn2 || ruleIgn2 != "passthrough" {
		t.Fatalf("expected tenant_id to be ignored, got mask=%v rule=%s", maskIgn2, ruleIgn2)
	}

	// Test unconfigured column in configured table
	_, maskUnconf := cfg.GetColumnRule("users", "unconfigured_field")
	if maskUnconf {
		t.Fatalf("expected unconfigured field not to be masked")
	}

	// Test orders table
	ruleCard, maskCard := cfg.GetColumnRule("orders", "credit_card")
	if !maskCard || ruleCard != "faker.credit_card(brand=visa)" {
		t.Fatalf("expected credit_card rule, got %s", ruleCard)
	}
}

func TestConfigValidationFailure(t *testing.T) {
	invalidYAML := `
version: "1.0"
tables:
  users:
    columns:
      email: "faker.nonexistent_strategy_123"
`
	_, err := LoadFromBytes([]byte(invalidYAML))
	if err == nil {
		t.Fatalf("expected validation error for nonexistent rule, got nil")
	}
}
