package stream

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alexandrmotologa/maskforge/internal/config"
	"github.com/alexandrmotologa/maskforge/internal/faker"
)

func TestPGCopyStreaming(t *testing.T) {
	cfgYAML := `
version: "1.0"
salt: "test-copy-salt"
tables:
  users:
    columns:
      email: "faker.email"
      first_name: "faker.first_name"
      phone: "faker.phone_number"
`
	cfg, err := config.LoadFromBytes([]byte(cfgYAML))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	inputSQL := `-- Sample PostgreSQL dump
CREATE TABLE users (
    id integer NOT NULL,
    email character varying(255),
    first_name character varying(100),
    phone character varying(50)
);

COPY public.users (id, email, first_name, phone) FROM stdin;
1	real.alice@company.com	Alice	+12025550192
2	real.bob@company.com	Bob	\N
3	\N	Charlie	+40722123456
\.

ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
`

	inBuf := bytes.NewBufferString(inputSQL)
	var outBuf bytes.Buffer

	pipeline := NewPipeline(cfg)
	metrics, err := pipeline.ProcessStream(inBuf, &outBuf)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	outStr := outBuf.String()

	// Verify DDL preserved
	if !strings.Contains(outStr, "CREATE TABLE users") {
		t.Fatalf("CREATE TABLE statement missing from output")
	}
	if !strings.Contains(outStr, "ALTER TABLE users") {
		t.Fatalf("ALTER TABLE statement missing from output")
	}

	// Verify original sensitive emails are gone
	if strings.Contains(outStr, "real.alice@company.com") {
		t.Fatalf("original sensitive email real.alice@company.com was not masked!")
	}
	if strings.Contains(outStr, "real.bob@company.com") {
		t.Fatalf("original sensitive email real.bob@company.com was not masked!")
	}

	// Verify \N was preserved
	if !strings.Contains(outStr, "\\N") {
		t.Fatalf("expected \\N to be preserved in output")
	}

	// Verify end of copy block
	if !strings.Contains(outStr, "\\.") {
		t.Fatalf("expected \\. to be preserved at end of copy block")
	}

	if metrics.RowsMasked != 3 {
		t.Fatalf("expected 3 rows masked, got %d", metrics.RowsMasked)
	}
}

func TestInsertStatementStreaming(t *testing.T) {
	cfgYAML := `
version: "1.0"
salt: "test-insert-salt"
tables:
  customers:
    columns:
      email: "faker.email"
      full_name: "faker.full_name"
`
	cfg, err := config.LoadFromBytes([]byte(cfgYAML))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	mysqlInput := `INSERT INTO ` + "`" + `customers` + "`" + ` (` + "`" + `id` + "`" + `, ` + "`" + `email` + "`" + `, ` + "`" + `full_name` + "`" + `) VALUES (101, 'sensitive.customer@bank.com', 'Arthur Pendelton'), (102, 'secret.director@gov.org', 'Morgan Le Fay');`

	inBuf := bytes.NewBufferString(mysqlInput + "\n")
	var outBuf bytes.Buffer

	pipeline := NewPipeline(cfg)
	metrics, err := pipeline.ProcessStream(inBuf, &outBuf)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	outStr := outBuf.String()

	if strings.Contains(outStr, "sensitive.customer@bank.com") {
		t.Fatalf("sensitive customer email was not masked")
	}
	if strings.Contains(outStr, "secret.director@gov.org") {
		t.Fatalf("secret director email was not masked")
	}
	if strings.Contains(outStr, "Arthur Pendelton") {
		t.Fatalf("original name was not masked")
	}

	if metrics.RowsMasked < 1 {
		t.Fatalf("expected at least 1 statement masked, got %d", metrics.RowsMasked)
	}
}

func TestSQLValuesTupleParsing(t *testing.T) {
	tuple := `(1, 'John O\'Connor', "some \"data\"", 45.6, NULL, 'line1\nline2')`
	values := ParseSQLValuesTuple(tuple)

	if len(values) != 6 {
		t.Fatalf("expected 6 values, got %d", len(values))
	}

	if values[0].Unquoted != "1" || values[0].IsString {
		t.Fatalf("expected 1 not string, got %+v", values[0])
	}
	if values[1].Unquoted != "John O'Connor" || !values[1].IsString {
		t.Fatalf("expected John O'Connor string, got %+v", values[1])
	}
	if !values[4].IsNull {
		t.Fatalf("expected NULL for index 4, got %+v", values[4])
	}
}

func TestExtractValuesTuplesMulti(t *testing.T) {
	section := `(1, 'A', 'foo(bar)'), (2, 'B', 'baz')`
	tuples := ExtractValuesTuples(section)

	if len(tuples) != 2 {
		t.Fatalf("expected 2 tuples, got %d", len(tuples))
	}
	if tuples[0] != `(1, 'A', 'foo(bar)')` {
		t.Fatalf("unexpected tuple 0: %s", tuples[0])
	}
	if tuples[1] != `(2, 'B', 'baz')` {
		t.Fatalf("unexpected tuple 1: %s", tuples[1])
	}
}

func TestFakerFormatPreservingPreservesUnchangedWhenUnconfigured(t *testing.T) {
	cfg := config.DefaultConfig()
	pipeline := NewPipeline(cfg)

	input := "SELECT * FROM settings;\n"
	var out bytes.Buffer
	_, err := pipeline.ProcessStream(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("process stream failed: %v", err)
	}

	if out.String() != input {
		t.Fatalf("expected identical passthrough, got %q", out.String())
	}
}

func TestDeterministicOutputConsistency(t *testing.T) {
	cfgYAML := `
version: "1.0"
salt: "consistent-seed"
tables:
  users:
    columns:
      email: "faker.email"
`
	cfg, _ := config.LoadFromBytes([]byte(cfgYAML))
	reg := faker.NewRegistry()

	ctx1 := faker.TransformContext{
		Salt:       cfg.Salt,
		Table:      "users",
		Column:     "email",
		RawValue:   "alice@example.com",
		RuleString: "faker.email",
	}
	res1 := reg.Transform(ctx1)
	res2 := reg.Transform(ctx1)

	if res1 != res2 {
		t.Fatalf("inconsistent deterministic transformation: %q vs %q", res1, res2)
	}
}

func TestConditionalMasking(t *testing.T) {
	cfgYAML := `
version: "1.0"
salt: "condition-salt"
tables:
  users:
    columns:
      email: "faker.email"
    conditions:
      email: "role != 'admin'"
`
	cfg, err := config.LoadFromBytes([]byte(cfgYAML))
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	inputSQL := `COPY public.users (id, role, email) FROM stdin;
1	admin	superadmin@internal-corp.com
2	member	customer.john@gmail.com
\.
`
	inBuf := bytes.NewBufferString(inputSQL)
	var outBuf bytes.Buffer

	pipeline := NewPipeline(cfg)
	_, err = pipeline.ProcessStream(inBuf, &outBuf)
	if err != nil {
		t.Fatalf("streaming failed: %v", err)
	}

	outStr := outBuf.String()

	// Admin email MUST be preserved
	if !strings.Contains(outStr, "superadmin@internal-corp.com") {
		t.Fatalf("expected admin email to be preserved due to condition role != 'admin', got: %s", outStr)
	}

	// Member email MUST be masked
	if strings.Contains(outStr, "customer.john@gmail.com") {
		t.Fatalf("expected member email to be masked, got: %s", outStr)
	}
}

func TestSamplingLimit(t *testing.T) {
	cfgYAML := `
version: "1.0"
salt: "sample-salt"
tables:
  logs:
    sample_limit: 2
`
	cfg, err := config.LoadFromBytes([]byte(cfgYAML))
	if err != nil {
		t.Fatalf("config error: %v", err)
	}

	inputSQL := `COPY public.logs (id, msg) FROM stdin;
1	line 1
2	line 2
3	line 3
4	line 4
5	line 5
\.
`
	inBuf := bytes.NewBufferString(inputSQL)
	var outBuf bytes.Buffer

	pipeline := NewPipeline(cfg)
	metrics, err := pipeline.ProcessStream(inBuf, &outBuf)
	if err != nil {
		t.Fatalf("process error: %v", err)
	}

	if metrics.RowsMasked != 2 {
		t.Fatalf("expected exactly 2 rows masked/written due to sample_limit: 2, got %d", metrics.RowsMasked)
	}
	if metrics.RowsSkipped != 3 {
		t.Fatalf("expected 3 rows skipped, got %d", metrics.RowsSkipped)
	}
}
