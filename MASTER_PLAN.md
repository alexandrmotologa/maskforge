# MaskForge: Master Implementation Plan

> Production Database Anonymizer, Sensitive PII Masker and Safe Staging Dump Generator for PostgreSQL and MySQL.
> High-speed streaming database anonymizer and privacy compliance engine. Pipes directly between pg_dump and psql without writing unmasked sensitive data to disk, automatically detects PII columns (names, emails, IBANs, cards, phone numbers), deterministically masks values to preserve foreign key relational integrity, and generates safe staging environments.

---

## 1. Project Overview and Problem Statement

Engineering teams need realistic database environments to debug production issues, verify performance characteristics, and run end-to-end integration tests locally. Using raw production dumps creates substantial privacy and security liabilities:
1. Regulatory Violations (GDPR, HIPAA, SOC 2, PCI-DSS): Storing unencrypted customer Personally Identifiable Information (PII) on developer laptops or unhardened staging environments exposes organizations to severe penalties and data breach liability.
2. Broken Relational Integrity: Naive masking scripts replace values with random strings without tracking seeds, immediately breaking foreign key constraints across tables (e.g. changing `user_id` in `users` breaks joins in `orders` and `invoices`).
3. Heavy Intermediate Disk Dumps: Traditional workflows dump databases to disk, run batch Python update scripts for hours, and re-export, exposing raw sensitive data on intermediate disk storage.
4. Validation Crashes: Replacing an email with random text or a credit card token with arbitrary digits crashes downstream application validations (e.g. failing email regex or Luhn checksums).

MaskForge resolves these issues through a high-speed, format-preserving streaming anonymizer:
- Zero-Disk Streaming Pipe: Reads SQL copy or insert streams directly from standard input and streams masked SQL to standard output:
  `pg_dump prod_db | maskforge --config mask.yaml | psql staging_db`. Unmasked bytes never touch disk.
- Auto-Discovery PII Scanner: Scans schemas and sample records, automatically identifying sensitive columns using column name patterns and regex content heuristics (names, emails, SSN, IBAN, credit cards, IP addresses, JWT tokens).
- Deterministic Faker Engine: Uses salted HMAC-SHA256 seeds to ensure referential consistency: an email or ID masked in one table is pseudonymized identically across all relational tables.
- Format-Preserving Transformations: Guarantees that masked emails remain valid RFC 5322 addresses, phone numbers keep valid international formats, and test credit card numbers satisfy the Luhn algorithm.
- Interactive Web Studio: Visual studio running on port 9880 to inspect database schemas, review detected PII classifications, preview before and after masked rows, and export clean `mask.yaml` configurations.

```
+-----------------------------------------------------------------------------------+
|                                     MASKFORGE                                     |
+-----------------------------------------+-----------------------------------------+
                                          |
        +---------------------------------+---------------------------------+
        |                                                                   |
+-------v-------------------------+                 +-----------------------v-------+
|    Streaming SQL Parser Engine  |                 |    Auto-Discovery PII Scanner |
|  * pg_dump COPY / INSERT Stream |                 |  * Column Name Heuristics     |
|  * MySQL mysqldump Format       |                 |  * Regex Value Matching       |
|  * Sub-Millisecond Line Tokenize|                 |  * Auto-Generates mask.yaml   |
+-------+-------------------------+                 +-----------------------+-------+
        |                                                                   |
        +---------------------------------+---------------------------------+
                                          |
                        +-----------------v-----------------+
                        |    Deterministic Faker Engine     |
                        |  * Salted HMAC Seed Mapping       |
                        |  * Foreign Key Referential Match  |
                        |  * Preserves Relational Joins     |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Format-Preserving Maskers       |
                        |  * Valid RFC 5322 Emails          |
                        |  * Luhn-Valid Test Card Numbers   |
                        |  * Phone, IBAN, SSN, Street Addr  |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |   Interactive Web Studio & CLI    |
                        |  * Visual Schema & PII Tree       |
                        |  * Live Before / After Row Diff   |
                        |  * Zero-Disk CLI (Port 9880)      |
                        +-----------------------------------+
```

---

## 2. Directory Structure

```
maskforge/
├── cmd/
│   └── maskforge/
│       └── main.go               # Unified CLI: run, inspect, studio, benchmark, version
├── internal/
│   ├── stream/
│   │   ├── pg_copy_reader.go     # PostgreSQL COPY stdin streaming parser
│   │   ├── pg_insert_reader.go   # PostgreSQL INSERT statement parser
│   │   ├── mysql_reader.go       # MySQL / MariaDB extended INSERT parser
│   │   ├── line_writer.go        # High-throughput buffered writer
│   │   └── pipeline.go           # Pipeline coordinator linking stream to faker
│   ├── discovery/
│   │   ├── pii_rules.go          # Regex matchers and classifications
│   │   ├── schema_scanner.go     # Schema and dump heuristic scanner
│   │   └── config_generator.go   # Generates baseline mask.yaml from scanned findings
│   ├── faker/
│   │   ├── deterministic.go      # Salted HMAC-SHA256 deterministic PRNG
│   │   ├── registry.go           # Transformer registry and dispatcher
│   │   ├── email.go              # RFC 5322 format-preserving email transformer
│   │   ├── names.go              # First, last, and full name dictionary generators
│   │   ├── phone.go              # International E.164 phone synthesizer
│   │   ├── finance.go            # Luhn-valid test cards and ISO 7064 IBANs
│   │   ├── identity.go           # SSN and national ID generators
│   │   ├── address.go            # Street, city, postal code, and country
│   │   ├── network.go            # Synthetic private and documentation IPs
│   │   ├── date.go               # Date of birth preserving realistic age bands
│   │   ├── hash.go               # Bcrypt staging passwords, SHA256, MD5
│   │   └── text.go               # Partial masking, redaction, and strings
│   ├── config/
│   │   ├── schema.go             # Configuration data models
│   │   └── loader.go             # YAML parser, validator, and default applicator
│   └── server/
│       ├── api.go                # REST API endpoints for studio
│       ├── embedded.go           # go:embed static asset handler
│       └── studio_service.go     # Studio backend engine (sample parser, diff runner)
├── web/
│   ├── src/
│   │   ├── components/
│   │   │   ├── SchemaTableView.tsx   # Table inventory and risk badges
│   │   │   ├── MaskRuleEditor.tsx    # Interactive rule and strategy selector
│   │   │   ├── RowDiffPreview.tsx    # Side-by-side original vs masked diff view
│   │   │   ├── ReferentialTest.tsx   # Cross-table foreign key consistency checker
│   │   │   ├── LiveSimulator.tsx     # Direct SQL snippet masking simulator
│   │   │   └── ConfigExportModal.tsx # YAML export and CLI command generator
│   │   ├── hooks/
│   │   │   └── useMaskSession.ts     # Client state management
│   │   ├── App.tsx                   # Main studio layout
│   │   ├── index.css                 # Styling and theme tokens
│   │   └── main.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── examples/
│   ├── mask.yaml                     # Sample configuration
│   ├── postgres_dump.sql             # Sample PostgreSQL test dump
│   └── mysql_dump.sql                # Sample MySQL test dump
├── docs/
│   ├── ARCHITECTURE.md               # Technical architecture and stream pipeline
│   ├── CONFIGURATION.md              # Configuration reference and rule syntax
│   ├── PERFORMANCE.md                # Benchmark analysis and memory profiling
│   └── SECURITY_COMPLIANCE.md        # GDPR, HIPAA, and PCI-DSS compliance details
├── go.mod
├── go.sum
└── README.md
```

---

## 3. Implementation Tasks and Subtasks

### Task 1: Scaffolding, Go Module & Git Setup
- Subtask 1.1: Initialize Go module `github.com/alexandrmotologa/maskforge` (Go 1.23+).
- Subtask 1.2: Set up dependencies (`gopkg.in/yaml.v3`, `github.com/spf13/cobra` or standard flag/cli package).
- Subtask 1.3: Set up `.gitignore`, license, and initial repository hygiene.

### Task 2: Core Deterministic Faker Engine
- Subtask 2.1: Salted HMAC-SHA256 deterministic generator (`internal/faker/deterministic.go`).
- Subtask 2.2: Name generators (`internal/faker/names.go`): first name, last name, full name.
- Subtask 2.3: Email generator (`internal/faker/email.go`): RFC 5322 compliance, preserve domain option.
- Subtask 2.4: Phone synthesizer (`internal/faker/phone.go`): country-coded phone numbers.
- Subtask 2.5: Financial generators (`internal/faker/finance.go`): Luhn mod-10 cards, ISO 7064 IBANs.
- Subtask 2.6: Identity, address, network, and text transformers (`identity.go`, `address.go`, `network.go`, `hash.go`, `text.go`).
- Subtask 2.7: Transformer registry and dispatcher (`registry.go`) supporting parameters and fallbacks.
- Subtask 2.8: Comprehensive unit tests asserting deterministic seed stability and referential consistency.

### Task 3: Declarative Configuration System
- Subtask 3.1: Data structures for `mask.yaml` (`internal/config/schema.go`).
- Subtask 3.2: YAML loader and validator (`internal/config/loader.go`).
- Subtask 3.3: Default rule handling, table wildcard matching, and column resolution.
- Subtask 3.4: Unit tests for valid and invalid configurations.

### Task 4: High-Speed Streaming SQL Parser
- Subtask 4.1: PostgreSQL COPY parser (`internal/stream/pg_copy_reader.go`):
  - Tokenize `COPY table (cols) FROM stdin;` statements.
  - Parse tab-separated columns without unbounded allocations.
  - Handle `\N` (NULL) and escape sequences.
  - Detect `\.` end of copy block.
- Subtask 4.2: PostgreSQL INSERT parser (`internal/stream/pg_insert_reader.go`):
  - Parse `INSERT INTO table (cols) VALUES (...), (...);`.
- Subtask 4.3: MySQL INSERT parser (`internal/stream/mysql_reader.go`):
  - Parse `INSERT INTO `table` VALUES (...);` with backticks and escaped quotes.
- Subtask 4.4: Line writer and streaming pipeline (`internal/stream/line_writer.go`, `pipeline.go`):
  - Buffer management (64KB - 256KB chunks).
  - Direct passthrough of DDL, comments, and non-configured tables.
  - Transformation of configured columns using the faker registry.
- Subtask 4.5: Unit and end-to-end streaming tests for PostgreSQL and MySQL streams.

### Task 5: Auto-Discovery PII Scanner
- Subtask 5.1: Regex rule definitions for column names and content patterns (`internal/discovery/pii_rules.go`).
- Subtask 5.2: Stream scanner inspecting table schemas and sample rows (`internal/discovery/schema_scanner.go`).
- Subtask 5.3: Automated configuration generator creating `mask.yaml` with recommended strategies (`internal/discovery/config_generator.go`).
- Subtask 5.4: Unit tests for detection precision and recall.

### Task 6: CLI Implementation
- Subtask 6.1: Command parsing (`cmd/maskforge/main.go`):
  - `run`: Streaming anonymization (`pg_dump | maskforge run | psql` or `-i input.sql -o output.sql`).
  - `inspect`: Schema and PII scan with terminal report and config generation.
  - `studio`: Launch the web studio on localhost.
  - `benchmark`: Built-in throughput benchmark.
  - `version`: Version and build metadata.
- Subtask 6.2: Piping ergonomics, graceful signal handling, and error reporting.

### Task 7: Web Studio Frontend (React 19)
- Subtask 7.1: Initialize React 19 + TypeScript + Vite + Tailwind CSS in `web/`.
- Subtask 7.2: Schema Table Explorer component with risk indicators.
- Subtask 7.3: Mask Rule Editor component with live strategy configuration.
- Subtask 7.4: Row Diff Preview component showing side-by-side original vs masked records.
- Subtask 7.5: Referential Integrity Tester confirming cross-table matching.
- Subtask 7.6: Live SQL Stream Simulator for interactive snippets.
- Subtask 7.7: Configuration export modal with 1-click YAML copy and CLI command.
- Subtask 7.8: Go embedding (`//go:embed`) and API server integration (`internal/server/`).

### Task 8: Documentation, Verification & Git Sync
- Subtask 8.1: Write technical documentation (`README.md`, `docs/ARCHITECTURE.md`, `docs/CONFIGURATION.md`, `docs/PERFORMANCE.md`, `docs/SECURITY_COMPLIANCE.md`).
- Subtask 8.2: Create sample datasets (`examples/postgres_dump.sql`, `examples/mysql_dump.sql`, `examples/mask.yaml`).
- Subtask 8.3: Run full test suite and benchmark suite to confirm throughput.
- Subtask 8.4: Commit and push clean working codebase to GitHub repository.
