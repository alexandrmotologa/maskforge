package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alexandrmotologa/maskforge/internal/config"
	"github.com/alexandrmotologa/maskforge/internal/discovery"
	"github.com/alexandrmotologa/maskforge/internal/faker"
	"github.com/alexandrmotologa/maskforge/internal/stream"
)

// TransformerMeta holds metadata about an available masking strategy.
type TransformerMeta struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Example     string   `json:"example"`
	Parameters  []string `json:"parameters"`
}

var transformerCatalog = []TransformerMeta{
	{
		Name:        "faker.email",
		Category:    "Identity",
		Description: "Generates valid RFC 5322 synthetic emails, with optional domain preservation.",
		Example:     "faker.email(preserve_domain=true)",
		Parameters:  []string{"preserve_domain", "domain"},
	},
	{
		Name:        "faker.first_name",
		Category:    "Identity",
		Description: "Generates realistic international given names.",
		Example:     "faker.first_name",
		Parameters:  []string{},
	},
	{
		Name:        "faker.last_name",
		Category:    "Identity",
		Description: "Generates realistic international surnames.",
		Example:     "faker.last_name",
		Parameters:  []string{},
	},
	{
		Name:        "faker.full_name",
		Category:    "Identity",
		Description: "Generates combined first and last names.",
		Example:     "faker.full_name",
		Parameters:  []string{},
	},
	{
		Name:        "faker.phone_number",
		Category:    "Contact",
		Description: "Generates valid E.164 phone numbers with country prefixes.",
		Example:     "faker.phone_number(prefix=+40)",
		Parameters:  []string{"prefix"},
	},
	{
		Name:        "faker.credit_card",
		Category:    "Financial",
		Description: "Generates 16-digit credit cards passing the Luhn mod-10 algorithm.",
		Example:     "faker.credit_card(brand=visa)",
		Parameters:  []string{"brand"},
	},
	{
		Name:        "faker.iban",
		Category:    "Financial",
		Description: "Generates valid IBAN with ISO 7064 Mod-97 check digits.",
		Example:     "faker.iban(country=RO)",
		Parameters:  []string{"country"},
	},
	{
		Name:        "faker.ssn",
		Category:    "Government ID",
		Description: "Generates valid US Social Security Numbers.",
		Example:     "faker.ssn",
		Parameters:  []string{},
	},
	{
		Name:        "faker.cnp",
		Category:    "Government ID",
		Description: "Generates Romanian Personal Numeric Codes (CNP) with valid checksum.",
		Example:     "faker.cnp",
		Parameters:  []string{},
	},
	{
		Name:        "faker.street_address",
		Category:    "Location",
		Description: "Generates realistic street addresses.",
		Example:     "faker.street_address",
		Parameters:  []string{},
	},
	{
		Name:        "faker.city",
		Category:    "Location",
		Description: "Generates realistic city names.",
		Example:     "faker.city",
		Parameters:  []string{},
	},
	{
		Name:        "faker.country",
		Category:    "Location",
		Description: "Generates realistic country names.",
		Example:     "faker.country",
		Parameters:  []string{},
	},
	{
		Name:        "faker.zip_code",
		Category:    "Location",
		Description: "Generates 5-digit postal codes.",
		Example:     "faker.zip_code",
		Parameters:  []string{},
	},
	{
		Name:        "faker.ipv4",
		Category:    "Network",
		Description: "Generates safe RFC 5737 or RFC 1918 IPv4 addresses.",
		Example:     "faker.ipv4(mode=private)",
		Parameters:  []string{"mode"},
	},
	{
		Name:        "faker.ipv6",
		Category:    "Network",
		Description: "Generates documentation RFC 3849 IPv6 addresses.",
		Example:     "faker.ipv6",
		Parameters:  []string{},
	},
	{
		Name:        "faker.uuid",
		Category:    "Identifiers",
		Description: "Generates deterministic UUIDv4 identifiers.",
		Example:     "faker.uuid",
		Parameters:  []string{},
	},
	{
		Name:        "faker.date_of_birth",
		Category:    "Personal",
		Description: "Generates birthdates keeping realistic adult age distribution.",
		Example:     "faker.date_of_birth(min_age=21, max_age=65)",
		Parameters:  []string{"min_age", "max_age"},
	},
	{
		Name:        "date.jitter",
		Category:    "Personal",
		Description: "Offsets date by +/- N days, preserving seasonal distribution.",
		Example:     "date.jitter(days=14)",
		Parameters:  []string{"days"},
	},
	{
		Name:        "hash.constant",
		Category:    "Security",
		Description: "Replaces secret with constant value (e.g. universal staging password).",
		Example:     "hash.constant('$2a$10$7EqJtq98hPqEX7fNZaFWoOhiMeqB7v1G8eM6Yt1PkW86i7D4rTfe6')",
		Parameters:  []string{"value"},
	},
	{
		Name:        "hash.sha256",
		Category:    "Security",
		Description: "Produces deterministic salted SHA256 hex string.",
		Example:     "hash.sha256",
		Parameters:  []string{},
	},
	{
		Name:        "mask.partial",
		Category:    "Obfuscation",
		Description: "Preserves prefix and suffix characters, masking the center with *.",
		Example:     "mask.partial(prefix=4, suffix=4, mask_char=*)",
		Parameters:  []string{"prefix", "suffix", "mask_char"},
	},
	{
		Name:        "mask.redact",
		Category:    "Obfuscation",
		Description: "Replaces value with fixed redaction string [REDACTED].",
		Example:     "mask.redact(label=[REDACTED])",
		Parameters:  []string{"label"},
	},
	{
		Name:        "passthrough",
		Category:    "System",
		Description: "Leaves original value unmasked and unchanged.",
		Example:     "passthrough",
		Parameters:  []string{},
	},
}

// Server handles API routes and web studio interactions.
type Server struct {
	mux *http.ServeMux
	reg *faker.Registry
}

// NewServer initializes the API server.
func NewServer() *Server {
	s := &Server{
		mux: http.NewServeMux(),
		reg: faker.NewRegistry(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/transformers", s.handleTransformers)
	s.mux.HandleFunc("/api/scan", s.handleScan)
	s.mux.HandleFunc("/api/transform/preview", s.handlePreview)
	s.mux.HandleFunc("/api/referential/test", s.handleReferentialTest)
	s.mux.HandleFunc("/api/stream/simulate", s.handleStreamSimulate)
	s.mux.HandleFunc("/api/config/export", s.handleConfigExport)
	s.mux.HandleFunc("/api/verify", s.handleVerify)
	s.mux.HandleFunc("/api/report/html", s.handleReportHTML)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "1.0.0",
		"service": "maskforge-studio",
	})
}

func (s *Server) handleTransformers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, transformerCatalog)
}

type ScanRequest struct {
	SQL        string `json:"sql"`
	MaxSamples int    `json:"max_samples"`
	Salt       string `json:"salt"`
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	scanner := discovery.NewSchemaScanner(req.Salt)
	report, err := scanner.ScanReader(strings.NewReader(req.SQL), req.MaxSamples)
	if err != nil {
		http.Error(w, "Scan failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

type PreviewRequest struct {
	Salt     string `json:"salt"`
	Table    string `json:"table"`
	Column   string `json:"column"`
	RawValue string `json:"raw_value"`
	Rule     string `json:"rule"`
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := faker.TransformContext{
		Salt:       req.Salt,
		Table:      req.Table,
		Column:     req.Column,
		RawValue:   req.RawValue,
		RuleString: req.Rule,
	}

	masked := s.reg.Transform(ctx)
	writeJSON(w, http.StatusOK, map[string]string{
		"raw_value":    req.RawValue,
		"masked_value": masked,
	})
}

type ReferentialTestRequest struct {
	Salt     string `json:"salt"`
	RawValue string `json:"raw_value"`
	Rule     string `json:"rule"`
	TableA   string `json:"table_a"`
	ColumnA  string `json:"column_a"`
	TableB   string `json:"table_b"`
	ColumnB  string `json:"column_b"`
}

func (s *Server) handleReferentialTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReferentialTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctxA := faker.TransformContext{
		Salt:       req.Salt,
		Table:      req.TableA,
		Column:     req.ColumnA,
		RawValue:   req.RawValue,
		RuleString: req.Rule,
	}
	ctxB := faker.TransformContext{
		Salt:       req.Salt,
		Table:      req.TableB,
		Column:     req.ColumnB,
		RawValue:   req.RawValue,
		RuleString: req.Rule,
	}

	maskedA := s.reg.Transform(ctxA)
	maskedB := s.reg.Transform(ctxB)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"raw_value": req.RawValue,
		"masked_a":  maskedA,
		"masked_b":  maskedB,
		"match":     maskedA == maskedB,
	})
}

type StreamSimulateRequest struct {
	SQL        string `json:"sql"`
	ConfigYAML string `json:"config_yaml"`
}

func (s *Server) handleStreamSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StreamSimulateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	cfg, err := config.LoadFromBytes([]byte(req.ConfigYAML))
	if err != nil {
		http.Error(w, "Invalid configuration YAML: "+err.Error(), http.StatusBadRequest)
		return
	}

	pipeline := stream.NewPipeline(cfg)
	var outBuf bytes.Buffer
	metrics, err := pipeline.ProcessStream(strings.NewReader(req.SQL), &outBuf)
	if err != nil {
		http.Error(w, "Streaming error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"masked_sql": outBuf.String(),
		"metrics":    metrics,
	})
}

type ConfigExportRequest struct {
	Config *config.Config `json:"config"`
}

func (s *Server) handleConfigExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfigExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Config == nil {
		http.Error(w, "Config cannot be nil", http.StatusBadRequest)
		return
	}

	yamlBytes, err := req.Config.ToYAML()
	if err != nil {
		http.Error(w, "Failed to serialize YAML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"yaml": string(yamlBytes),
	})
}

type VerifyRequest struct {
	SQL                 string   `json:"sql"`
	AllowDomains        []string `json:"allow_domains"`
	AllowSyntheticCards bool     `json:"allow_synthetic_cards"`
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	opts := discovery.VerifyOptions{
		AllowedDomains:      req.AllowDomains,
		AllowSyntheticCards: req.AllowSyntheticCards,
	}

	report, err := discovery.VerifySQLDump(strings.NewReader(req.SQL), opts)
	if err != nil {
		http.Error(w, "Verification failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

type ReportHTMLRequest struct {
	Report       *discovery.ScanReport `json:"report"`
	DatabaseName string                `json:"database_name"`
}

func (s *Server) handleReportHTML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReportHTMLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Report == nil {
		http.Error(w, "Report cannot be nil", http.StatusBadRequest)
		return
	}

	html := discovery.GenerateHTMLComplianceReport(req.Report, req.DatabaseName)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
