# MaskForge

Production database anonymizer, sensitive PII masker, and safe staging dump generator for PostgreSQL and MySQL.

MaskForge operates as a high-speed streaming UNIX pipe between database dump utilities and import tools. It automatically detects Personally Identifiable Information (PII), generates realistic synthetic data, and preserves relational integrity across foreign keys using deterministic HMAC-SHA256 seed mapping.

```bash
pg_dump prod_db | maskforge run --config mask.yaml | psql staging_db
```

Zero unmasked bytes ever touch disk.

---

## Key Capabilities

- **Zero-Disk Streaming Filter**: Reads standard SQL COPY and INSERT streams from STDIN and emits masked SQL to STDOUT with 256 KB streaming buffers.
- **Automated PII Discovery**: Scans schemas and sample records, classifying sensitive columns against 50+ heuristics across GDPR, HIPAA, PCI-DSS, and SOC 2 requirements.
- **Deep JSON / JSONB Masking**: Traverses nested JSON structures (`faker.json(paths)`) to anonymize sensitive attributes inside payloads without breaking valid JSON syntax.
- **Conditional Masking Rules**: Preserves internal admin accounts, test users, or specific roles while anonymizing production customer data (`when: "role != 'admin'"`).
- **Zero-Leak Security Verifier**: Post-masking adversarial scanner (`maskforge verify`) auditing outputs to certify zero sensitive PII leaks for SOC 2 and GDPR auditors.
- **Database Subsetting & Sampling**: Downsamples multi-hundred gigabyte dumps to lightweight staging environments (`--sample-rate 0.1` or `--limit-rows 1000`).
- **Deterministic Referential Integrity**: Maps identical foreign key values across relational tables to the same synthetic pseudonyms without breaking database constraints.
- **Format-Preserving Transformers**: Synthetic emails satisfy RFC 5322, credit cards satisfy Luhn mod-10, IBANs satisfy ISO 7064 check digits, and national IDs maintain format standards.
- **Statistical Numeric Jitter**: Adds configurable variance to financial totals, salaries, and order amounts while maintaining statistical distributions.
- **Interactive Web Studio**: Built-in React 19 web interface running on port 9880 to audit schemas, preview before and after row diffs, and test cross-table join integrity.
- **High Throughput**: Processes over 140,000 rows per second with under 30 MB resident memory.

---

## Architecture

MaskForge processes SQL dumps as a streaming pipeline:

```
[ pg_dump / mysqldump ]
          │ (stdout pipe)
          ▼
   [ os.Stdin ]
          │
          ▼
┌────────────────────────────────────────────────────────┐
│ MaskForge Stream Engine                                │
│                                                        │
│  ┌───────────────────────┐   ┌──────────────────────┐  │
│  │ bufio.Reader (256 KB) │   │ Dialect Discriminator│  │
│  └──────────┬────────────┘   └──────────┬───────────┘  │
│             │                           │              │
│             ▼                           ▼              │
│  ┌───────────────────────┐   ┌──────────────────────┐  │
│  │ Statement Dispatcher  │──▶│ COPY / INSERT Parser │  │
│  └───────────────────────┘   └──────────┬───────────┘  │
│                                         │              │
│                                         ▼              │
│                              ┌──────────────────────┐  │
│                              │ Conditional Evaluator│  │
│                              │ (when expressions)   │  │
│                              └──────────┬───────────┘  │
│                                         │              │
│                                         ▼              │
│                              ┌──────────────────────┐  │
│                              │ Deterministic Faker  │  │
│                              │ HMAC-SHA256 PRNG     │  │
│                              └──────────┬───────────┘  │
│                                         │              │
│                                         ▼              │
│                              ┌──────────────────────┐  │
│                              │ Format Preservers    │  │
│                              │ RFC 5322, Luhn Mod10 │  │
│                              └──────────┬───────────┘  │
│                                         │              │
│  ┌───────────────────────┐              │              │
│  │ bufio.Writer (256 KB) │◀─────────────┘              │
│  └──────────┬────────────┘                             │
└─────────────┼──────────────────────────────────────────┘
              │ (stdout pipe)
              ▼
        [ os.Stdout ]
              │
              ▼
      [ psql / mysql ]
```

1. **Dialect Discriminator**: Identifies PostgreSQL `COPY ... FROM stdin;` blocks, PostgreSQL `INSERT INTO` statements, and MySQL extended multi-row `INSERT` statements.
2. **Passthrough Engine**: DDL statements (`CREATE TABLE`, `ALTER TABLE`, `CREATE INDEX`), comments, transaction boundaries, and sequence values stream directly to output with zero modification overhead.
3. **Deterministic Seed Generator**: Inputs `(Salt, Context, RawValue)` derive a stable 64-bit state for reproducible pseudonym generation.

---

## Installation

### From Source (Go 1.23+)

```bash
git clone https://github.com/alexandrmotologa/maskforge.git
cd maskforge
go build -o maskforge ./cmd/maskforge
```

### Pre-Built Binaries

Download the appropriate binary for your operating system from the GitHub Releases page:
- Linux (x86_64, aarch64)
- macOS (Apple Silicon, Intel)
- Windows (x86_64)

---

## Quickstart

### 1. Audit Database Dump and Generate Configuration

Scan an existing SQL dump file to detect sensitive columns, export a configuration, or generate a standalone HTML compliance report:

```bash
maskforge inspect prod_dump.sql --generate-config mask.yaml --export-report compliance_report.html
```

Output:
```
========================================================================================
                             MASKFORGE PII AUDIT REPORT                                
========================================================================================
Total Tables:     2
Total Columns:    15
PII Columns:      10
Estimated Risk:   CRITICAL (Critical: 3, High: 3, Medium: 4, Low: 0)

TABLE: users (Rows: 5)
----------------------------------------------------------------------------------------
COLUMN                   TYPE                 RISK       RECOMMENDED RULE       COMPLIANCE
----------------------------------------------------------------------------------------
id                       Non-Sensitive / ID   LOW        passthrough            -
email                    Email Address        HIGH       faker.email            GDPR Art. 4
first_name               First Name           MEDIUM     faker.first_name       GDPR Art. 4
last_name                Last Name            MEDIUM     faker.last_name        GDPR Art. 4
phone_number             Phone Number         HIGH       faker.phone_number     GDPR Art. 4
ssn                      National ID / SSN    CRITICAL   faker.ssn              GDPR Art. 87
password_hash            Password / Secret    CRITICAL   hash.constant          SOC 2, ISO 27001
created_at               Non-Sensitive / ID   LOW        passthrough            -
```

### 2. Stream Anonymized Data directly into Staging

Pipe directly from `pg_dump` into `psql`:

```bash
pg_dump prod_db | maskforge run --config mask.yaml | psql staging_db
```

Or between files, optionally sampling 10% of rows:

```bash
maskforge run --config mask.yaml --input prod.sql --output staging.sql --sample-rate 0.1
```

### 3. Verify Dump for Zero Sensitive Data Leaks

Audit an already-masked SQL dump to certify that no live credit cards, production emails, or SSNs remain:

```bash
maskforge verify staging.sql --allow-test-cards
```

Output:
```
========================================================================================
                         MASKFORGE ZERO-LEAK AUDIT REPORT                               
========================================================================================
Lines Audited:    48
Compliance Score: 100 / 100

[PASSED] ZERO SENSITIVE PII LEAKS DETECTED.
Database dump is certified safe for staging and local development.
```

### 4. Launch Interactive Web Studio

Start the embedded visual studio:

```bash
maskforge studio --port 9880
```

Open `http://localhost:9880` in your browser to inspect table risk matrices, apply 1-click compliance presets (GDPR, PCI-DSS, HIPAA), test foreign key preservation, and run leak audits.

---

## Configuration Reference (`mask.yaml`)

```yaml
version: "1.0"
salt: "c8e2b9f10a4d3e71829304a5b6c7d8e9"

global:
  preserve_nulls: true
  ignore_columns:
    - "id"
    - "created_at"
    - "updated_at"

tables:
  users:
    columns:
      email: "faker.email"
      first_name: "faker.first_name"
      last_name: "faker.last_name"
      phone_number: "faker.phone_number(prefix=+40)"
      ssn: "faker.cnp"
      password_hash: "hash.constant"
    conditions:
      email: "role != 'admin' AND is_internal = false"
    ignore_columns:
      - "tenant_id"
    sample_limit: 5000

  events:
    columns:
      payload: "faker.json(customer.email=faker.email, customer.name=faker.full_name)"

  payments:
    columns:
      card_holder: "faker.full_name"
      card_number: "faker.credit_card(brand=visa)"
      iban: "faker.iban(country=RO)"
      billing_address: "faker.street_address"
      ip_address: "faker.ipv4(mode=private)"

  orders:
    columns:
      customer_email: "faker.email" # Matches users.email deterministically
      total_amount: "numeric.jitter(percent=10)"
```

### Available Transformers

| Category | Transformer | Parameters | Description |
|---|---|---|---|
| **Identity** | `faker.first_name` | - | Realistic given names |
| | `faker.last_name` | - | Realistic family names |
| | `faker.full_name` | - | Combined first and last name |
| | `faker.username` | - | Synthetic handle (`alex.smith42`) |
| **Contact** | `faker.email` | `preserve_domain=true` | RFC 5322 email address |
| | `faker.phone_number` | `prefix=+1`, `prefix=+40` | E.164 international phone number |
| **Finance** | `faker.credit_card` | `brand=visa`, `brand=mastercard` | Luhn mod-10 compliant test card number |
| | `faker.iban` | `country=RO`, `country=DE` | Valid IBAN with ISO 7064 check digits |
| | `faker.cvv` | `length=3` | Card verification code |
| | `numeric.jitter` | `percent=10` | Adds deterministic statistical variance |
| | `numeric.range` | `min=100`, `max=500` | Bounds numbers into range |
| **Government** | `faker.ssn` | - | US Social Security Number |
| | `faker.cnp` | - | Romanian Personal Numeric Code |
| | `faker.national_id` | `country=RO` | Generic national identifier |
| **Nested Data** | `faker.json` | `path1=rule1, path2=rule2` | Recursively masks target JSON/JSONB keys |
| **Location** | `faker.street_address` | - | Street name and number |
| | `faker.city` | - | City name |
| | `faker.country` | - | Country name |
| | `faker.zip_code` | - | Postal code |
| **Network** | `faker.ipv4` | `mode=private`, `mode=doc` | RFC 1918 or RFC 5737 IP address |
| | `faker.ipv6` | - | Documentation RFC 3849 IPv6 address |
| | `faker.uuid` | - | Valid RFC 4122 UUIDv4 |
| **Personal** | `faker.date_of_birth` | `min_age=18`, `max_age=65` | Birthdate in `YYYY-MM-DD` |
| | `date.jitter` | `days=14` | Shifts date within +/- N days |
| **Security** | `hash.constant` | `value="$2a$10$..."` | Default bcrypt hash for `password123` |
| | `hash.sha256` | - | Deterministic salted SHA-256 |
| **Obfuscation** | `mask.partial` | `prefix=4`, `suffix=4` | Preserves edges, masks center with `*` |
| | `mask.redact` | `label="[REDACTED]"` | Constant redaction label |
| | `passthrough` | - | Leaves value unchanged |

---

## Verification and Testing

Execute the test suite across all packages:

```bash
go test -v ./...
```

Run the built-in streaming benchmark:

```bash
maskforge benchmark --rows 100000
```

Benchmark output:
```
Benchmarking MaskForge streaming engine with 100000 synthetic rows...

Benchmark Results:
  Total Rows Masked: 100000
  Data Streamed:     8.98 MB
  Duration:          712ms
  Throughput:        140365 rows/sec
  Stream Bandwidth:  12.60 MB/sec
  RAM Allocated:     28.42 MB
```

---

## Documentation

- [Architecture & Data Flow](docs/ARCHITECTURE.md)
- [Configuration Reference](docs/CONFIGURATION.md)
- [Performance & Benchmarks](docs/PERFORMANCE.md)
- [Security & Compliance Mapping](docs/SECURITY_COMPLIANCE.md)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
