# MaskForge Security and Compliance Guide

Using raw production database dumps in local developer environments or staging clusters introduces substantial compliance and data breach liabilities. MaskForge is engineered to address regulatory requirements under GDPR, HIPAA, PCI-DSS, and SOC 2.

## Regulatory Mapping

### 1. General Data Protection Regulation (GDPR - EU 2016/679)
- **Article 4 (Pseudonymisation)**: Personal data must be processed such that it can no longer be attributed to a specific data subject without the use of additional information kept separately.
  - *MaskForge Implementation*: Salted HMAC-SHA256 irreversibly pseudonymizes user identifiers, emails, names, and phone numbers. The salt secret is stored outside the database.
- **Article 32 (Security of Processing)**: Requires pseudonymisation and encryption of personal data in non-production environments.
  - *MaskForge Implementation*: The streaming architecture ensures that unmasked production dumps are never stored on developer workstations or CI/CD runner disks.
- **Article 87 (National Identification Numbers)**: Specific restrictions on processing government identifiers.
  - *MaskForge Implementation*: Dedicated transformers for SSN and CNP replace genuine national IDs with algorithmically valid synthetic values.

### 2. Payment Card Industry Data Security Standard (PCI-DSS v4.0)
- **Requirement 3.4 (Primary Account Numbers)**: PAN must be rendered unreadable anywhere it is stored, including backups and test environments.
  - *MaskForge Implementation*: `faker.credit_card` replaces real customer card numbers with synthetic test PANs that satisfy the Luhn mod-10 algorithm. CVVs and expiration tokens are completely sanitized.

### 3. Health Insurance Portability and Accountability Act (HIPAA)
- **45 CFR § 164.514 (Safe Harbor Method)**: Requires removal of 18 specific identifiers from health datasets, including names, geographic subdivisions, dates directly related to an individual, telephone numbers, fax numbers, email addresses, and Social Security numbers.
  - *MaskForge Implementation*: Auto-discovery scanner flags all 18 HIPAA identifier types and assigns appropriate format-preserving transformers.

### 4. SOC 2 Type II
- **Trust Services Criteria (Confidentiality & Privacy)**: Testing environments must not store production secrets or customer credentials.
  - *MaskForge Implementation*: `hash.constant` replaces all production password hashes, session tokens, and API keys with uniform development credentials (e.g. standard bcrypt hash for `password123`).

## Salt Management and Security Best Practices

1. **Keep Salts Confidential**:
   The deterministic salt acts as the secret key. If an adversary obtains both the salt and the masked database, they can mount dictionary attacks against low-entropy values. Store `mask.yaml` or the salt string in a secrets manager (such as HashiCorp Vault or AWS Secrets Manager).

2. **Rotate Salts Across Environments**:
   Use different salts for different staging clusters or developer teams. This guarantees that pseudonyms in Staging Environment A cannot be correlated with pseudonyms in Staging Environment B.

3. **Verify Pipeline Exit Codes**:
   When using pipe mode (`pg_dump | maskforge | psql`), ensure shell option `set -o pipefail` is active in CI/CD scripts so failures in any stage halt the execution immediately.
