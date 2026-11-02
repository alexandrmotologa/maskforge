package faker

import (
	"fmt"
	"strings"
)

var defaultEmailDomains = []string{
	"example.com",
	"example.org",
	"example.net",
	"testmail.dev",
	"staging.internal",
	"sandbox.local",
}

// GenerateEmail synthesizes a valid RFC 5322 email.
// If preserveDomain is true and originalValue contains an '@' followed by a dot,
// the original domain is preserved while the local-part is anonymized.
func GenerateEmail(rng *DeterministicPRNG, originalValue string, preserveDomain bool, customDomain string) string {
	fn := strings.ToLower(GenerateFirstName(rng))
	ln := strings.ToLower(GenerateLastName(rng))
	code := rng.IntBetween(10, 999)
	localPart := fmt.Sprintf("%s.%s%d", fn, ln, code)

	domain := ""
	if customDomain != "" {
		domain = customDomain
	} else if preserveDomain && strings.Contains(originalValue, "@") {
		parts := strings.SplitN(originalValue, "@", 2)
		if len(parts) == 2 && strings.Contains(parts[1], ".") && !strings.ContainsAny(parts[1], " \t\r\n'\"") {
			domain = strings.ToLower(strings.TrimSpace(parts[1]))
		}
	}

	if domain == "" {
		domain = rng.PickString(defaultEmailDomains)
	}

	return fmt.Sprintf("%s@%s", localPart, domain)
}
