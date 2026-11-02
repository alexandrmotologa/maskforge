package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", resp["status"])
	}
}

func TestTransformersEndpoint(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest("GET", "/api/transformers", nil)
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var list []TransformerMeta
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to parse transformers list: %v", err)
	}
	if len(list) < 10 {
		t.Fatalf("expected at least 10 transformers, got %d", len(list))
	}
}

func TestPreviewEndpoint(t *testing.T) {
	srv := NewServer()
	body := map[string]string{
		"salt":      "preview-test-salt",
		"table":     "users",
		"column":    "email",
		"raw_value": "alice.wonderland@secret.org",
		"rule":      "faker.email",
	}
	jsonBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/transform/preview", bytes.NewReader(jsonBytes))
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp["masked_value"] == "" || resp["masked_value"] == "alice.wonderland@secret.org" {
		t.Fatalf("expected masked email, got %q", resp["masked_value"])
	}
}

func TestReferentialTestEndpoint(t *testing.T) {
	srv := NewServer()
	body := ReferentialTestRequest{
		Salt:     "shared-salt",
		RawValue: "cust_45890",
		Rule:     "faker.uuid",
		TableA:   "users",
		ColumnA:  "id",
		TableB:   "orders",
		ColumnB:  "customer_id",
	}
	jsonBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/referential/test", bytes.NewReader(jsonBytes))
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp["match"] != true {
		t.Fatalf("expected match to be true across tables, got %v", resp)
	}
	if resp["masked_a"] != resp["masked_b"] {
		t.Fatalf("expected masked_a == masked_b, got %v vs %v", resp["masked_a"], resp["masked_b"])
	}
}

func TestStreamSimulateEndpoint(t *testing.T) {
	srv := NewServer()
	body := StreamSimulateRequest{
		SQL: "INSERT INTO `users` (`id`, `email`) VALUES (1, 'john@doe.com');",
		ConfigYAML: `
version: "1.0"
salt: "test"
tables:
  users:
    columns:
      email: "faker.email"
`,
	}
	jsonBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/stream/simulate", bytes.NewReader(jsonBytes))
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	maskedSQL, ok := resp["masked_sql"].(string)
	if !ok || maskedSQL == "" {
		t.Fatalf("expected masked_sql string in response")
	}
	if bytes.Contains([]byte(maskedSQL), []byte("john@doe.com")) {
		t.Fatalf("expected email to be masked in simulated stream")
	}
}

func TestVerifyEndpoint(t *testing.T) {
	srv := NewServer()
	body := VerifyRequest{
		SQL: "COPY public.users (id, email) FROM stdin;\n1\tclean@example.com\n\\.\n",
	}
	jsonBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/verify", bytes.NewReader(jsonBytes))
	rec := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp["is_clean"] != true {
		t.Fatalf("expected is_clean true for clean test stream")
	}
}
