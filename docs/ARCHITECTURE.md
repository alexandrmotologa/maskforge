# MaskForge Architecture

## 1. System Design and Data Flow

MaskForge is engineered as a streaming UNIX filter operating between standard input and standard output. The engine never buffers full database dumps into memory or intermediate temporary files on disk.

```
+-------------------------------------------------------------------------------+
|                               Pipeline Topology                               |
+-------------------------------------------------------------------------------+

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

## 2. Core Components

### 2.1 Stream Parser (`internal/stream`)
- **PostgreSQL COPY Parser**: Recognizes `COPY [schema.]table (cols) FROM stdin;`. Rows are tab-delimited. The parser splits tabs in-place, transforms configured column indices, and writes directly to the 256 KB output buffer without object reallocation.
- **INSERT Parser**: Tokenizes single and multi-row value tuples (`(1, 'foo', ...), (2, 'bar', ...)`). Identifies string literals, numbers, boolean values, and NULL tokens while honoring escaped single quotes (`''`, `\'`) and backslashes.
- **DDL Passthrough**: Any line that is not a target COPY or INSERT record (e.g. `CREATE TABLE`, `ALTER TABLE`, comments, sequence settings) is passed through verbatim with zero transformation overhead.

### 2.2 Deterministic Faker Engine (`internal/faker`)
Relational databases enforce foreign key constraints across tables (e.g. `orders.customer_email` referencing `users.email`). Random replacement breaks these joins immediately.

MaskForge resolves this via a deterministic pseudorandom number generator:
1. Inputs: Salt string (from `mask.yaml`), Column context name, Raw sensitive value.
2. Hash: Computes `HMAC-SHA256(Salt, Context + \0 + RawValue)`.
3. Generator: Derives high-entropy 64-bit words to seed a PCG-style generator.
4. Property: Given the same salt and raw value, the output is guaranteed to be identical across all tables, rows, and runs.

### 2.3 Format-Preserving Transformers
- **Email (`faker.email`)**: Generates valid RFC 5322 email addresses. Supports preserving original corporate domains (`preserve_domain: true`) while pseudonymizing the mailbox part.
- **Credit Card (`faker.credit_card`)**: Produces 15 or 16 digit numbers matching Visa or Mastercard prefixes that strictly satisfy the Luhn mod-10 algorithm.
- **IBAN (`faker.iban`)**: Generates valid IBAN structures with ISO 7064 Mod-97 check digits for Romania, Germany, France, the United Kingdom, and generic SEPA countries.
- **National IDs (`faker.ssn`, `faker.cnp`)**: US Social Security Numbers avoiding SSA invalid ranges; Romanian CNP numbers conforming to official 12-digit weighted checksum calculation.
- **Staging Password Hashes (`hash.constant`)**: Replaces complex production password hashes with a static bcrypt hash for `password123`, allowing developers to authenticate into any test account in staging.

### 2.4 Auto-Discovery Scanner (`internal/discovery`)
- **Tier 1 Heuristic**: Inspects column names against 50+ regular expression patterns.
- **Tier 2 Sniffing**: Inspects sample row values to detect valid emails, Luhn cards, IBANs, and IPv4 addresses when column names are ambiguous.
- **Classification**: Assigns risk ratings (`CRITICAL`, `HIGH`, `MEDIUM`, `LOW`) and maps them to compliance standards (GDPR, HIPAA, PCI-DSS, SOC 2).

### 2.5 Web Studio Architecture (`internal/server` and `web/`)
- Backend runs an embedded Go HTTP service.
- Frontend is compiled with React 19, TypeScript, and Tailwind CSS.
- Static assets are embedded directly into the Go binary using Go 1.16+ `embed.FS`, producing a single self-contained executable.
