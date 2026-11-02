# MaskForge Configuration Reference

Configurations are declarative YAML files (`mask.yaml`) specifying tables, columns, transformer rules, sampling thresholds, and conditional expressions.

## Basic Structure

```yaml
version: "1.0"
salt: "your-deterministic-salt-secret-key"

global:
  preserve_nulls: true
  sample_rate: 0.1 # Retains 10% of rows globally
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
      ssn: "faker.ssn"
      password_hash: "hash.constant"
    conditions:
      email: "role != 'system_admin' AND is_internal = false"
    ignore_columns:
      - "tenant_id"
    sample_limit: 5000

  events:
    columns:
      payload: "faker.json(customer.email=faker.email, billing.card=faker.credit_card)"

  orders:
    columns:
      customer_email: "faker.email"
      shipping_address: "faker.street_address"
      card_number: "faker.credit_card(brand=visa)"
      total_amount: "numeric.jitter(percent=15)"
```

## Global Options

| Field | Type | Default | Description |
|---|---|---|---|
| `version` | string | `"1.0"` | Configuration schema version. |
| `salt` | string | Auto-generated | Secret key used for HMAC-SHA256 deterministic seeding. |
| `global.preserve_nulls` | boolean | `true` | When true, SQL `NULL` and `\N` tokens pass through unchanged. |
| `global.sample_rate` | float | `0.0` | Global row sampling fraction (0.01 to 1.0). |
| `global.default_rule` | string | `""` | Fallback rule applied to any column not matched by explicit configuration. |
| `global.ignore_tables` | list[string] | `[]` | Tables completely excluded from transformation. |
| `global.ignore_columns` | list[string] | `[]` | Column names ignored globally across all tables. |

## Table Options

| Field | Type | Description |
|---|---|---|
| `columns` | map[string]string | Maps column names to transformer rule strings. |
| `conditions` | map[string]string | Maps column names to conditional `when` expressions. |
| `ignore_columns` | list[string] | List of columns in this table to leave unchanged. |
| `sample_limit` | integer | Maximum number of rows to stream for this table. |
| `truncate` | boolean | If true, drops all data rows from this table. |

## Conditional Masking (`when`)

Conditional masking allows selective anonymization based on row values. For example, preserving internal employee or admin accounts while masking real customer data:

```yaml
tables:
  users:
    columns:
      email: "faker.email"
    conditions:
      email: "role != 'admin' AND is_active = true"
```

Supported operators:
- Equality: `=`, `!=`, `<>`
- Comparison: `>`, `<`, `>=`, `<=`
- Pattern matching: `LIKE '%pattern%'`, `NOT LIKE '%internal%'`
- Logical composition: `AND`, `OR`

## Nested JSON / JSONB Masking (`faker.json`)

To mask sensitive attributes inside JSON objects without breaking the outer JSON syntax:

```yaml
tables:
  webhook_events:
    columns:
      payload: "faker.json(customer.email=faker.email, customer.phone=faker.phone_number, card.number=faker.credit_card)"
```

Paths are evaluated recursively. Nested keys not targeted in the rule string remain intact.

## Available Transformers

### Identity and Names
- `faker.first_name`: Realistic international first names.
- `faker.last_name`: Realistic international last names.
- `faker.full_name`: Combined first and last name.
- `faker.username`: Synthetic handle (e.g. `alex.smith42`).

### Contact Information
- `faker.email`: Valid RFC 5322 synthetic email.
  - Parameters:
    - `preserve_domain=true`: Retains original corporate email domain while anonymizing local mailbox part.
    - `domain=test.com`: Forces specific synthetic domain.
- `faker.phone_number`: Valid international E.164 phone number.
  - Parameters:
    - `prefix=+1`: US country prefix.
    - `prefix=+40`: Romanian prefix (+40 7xx xxx xxx).
    - `prefix=+44`: UK prefix (+44 7xxx xxxxxx).
    - `prefix=+49`: German prefix (+49 15x xxxxxxx).

### Financial Data
- `faker.credit_card`: 16-digit credit card number satisfying Luhn mod-10 check.
  - Parameters:
    - `brand=visa`: Card starting with digit 4.
    - `brand=mastercard`: Card starting with 51-55.
    - `brand=amex`: 15-digit card starting with 34/37.
- `faker.cvv`: 3 or 4 digit card verification code.
- `faker.iban`: Valid IBAN format with ISO 7064 Mod-97 check digits.
  - Parameters:
    - `country=RO`: Romanian IBAN (24 chars).
    - `country=DE`: German IBAN (22 chars).
    - `country=FR`: French IBAN (27 chars).
    - `country=GB`: UK IBAN (22 chars).
- `numeric.jitter`: Adds +/- percent statistical variance to numeric columns.
  - Parameters:
    - `percent=15`: Variance percentage.
- `numeric.range`: Bounds numeric value into [min, max] range.
  - Parameters:
    - `min=100`, `max=500`.

### Government Identifiers
- `faker.ssn`: US Social Security Number (`AAA-GG-SSSS`).
- `faker.cnp`: Romanian Personal Numeric Code (CNP) with valid mod-11 check digit.
- `faker.national_id`: Generic format national identity number.

### Network and Identifiers
- `faker.ipv4`: Safe synthetic IPv4 address.
  - Parameters:
    - `mode=private`: RFC 1918 range (`10.x.x.x`).
    - `mode=doc`: RFC 5737 documentation blocks (`192.0.2.x`, `198.51.100.x`).
- `faker.ipv6`: Documentation RFC 3849 IPv6 address (`2001:0db8:...`).
- `faker.uuid`: Deterministic valid UUIDv4 string.

### Dates
- `faker.date_of_birth`: Realistic birthdate in `YYYY-MM-DD` format.
  - Parameters:
    - `min_age=21`, `max_age=65`: Custom age bounds.
- `date.jitter`: Shifts existing dates by +/- N days to preserve seasonality without exposing exact timestamps.
  - Parameters:
    - `days=14`: Maximum day offset.

### Passwords and Hashes
- `hash.constant`: Constant string replacement. Defaults to standard bcrypt hash for `password123` (`$2a$10$7EqJtq98hPqEX7fNZaFWoOhiMeqB7v1G8eM6Yt1PkW86i7D4rTfe6`).
  - Parameters:
    - `value="custom_string"`: Overrides default hash.
- `hash.sha256`: Salted SHA-256 hexadecimal output.
- `hash.md5`: Salted MD5 hexadecimal output.

### Obfuscation and Strings
- `mask.partial`: Keeps specified prefix and suffix characters, replacing central characters with mask character.
  - Parameters:
    - `prefix=4`, `suffix=4`, `mask_char=*`: e.g. `4111********1111`.
- `mask.redact`: Constant label (default `[REDACTED]`).
  - Parameters:
    - `label="CONFIDENTIAL"`
- `mask.digits`: Replaces digits with synthetic digits, preserving non-digit symbols.
