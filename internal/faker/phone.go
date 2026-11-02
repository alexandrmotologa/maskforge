package faker

import (
	"fmt"
	"strings"
)

var defaultCountryPrefixes = []string{
	"+1",  // US/Canada
	"+40", // Romania
	"+44", // UK
	"+49", // Germany
	"+33", // France
	"+34", // Spain
	"+39", // Italy
	"+31", // Netherlands
}

// GeneratePhoneNumber produces a valid E.164-style phone number.
// If countryPrefix is empty, it either extracts the prefix from originalValue if present,
// or chooses deterministically from common prefixes.
func GeneratePhoneNumber(rng *DeterministicPRNG, originalValue string, targetPrefix string) string {
	prefix := targetPrefix
	if prefix == "" {
		trimmed := strings.TrimSpace(originalValue)
		if strings.HasPrefix(trimmed, "+") {
			// Extract known prefix
			for _, cp := range defaultCountryPrefixes {
				if strings.HasPrefix(trimmed, cp) {
					prefix = cp
					break
				}
			}
		}
	}
	if prefix == "" {
		prefix = rng.PickString(defaultCountryPrefixes)
	}

	switch prefix {
	case "+1":
		// US format: +1 (200-999) (200-999) (0000-9999)
		area := rng.IntBetween(201, 989)
		exchange := rng.IntBetween(200, 999)
		subscriber := rng.IntBetween(1000, 9999)
		return fmt.Sprintf("+1%03d%03d%04d", area, exchange, subscriber)
	case "+40":
		// Romania: +40 7xx xxx xxx
		sub := rng.IntBetween(10, 89)
		tail := rng.IntBetween(100000, 999999)
		return fmt.Sprintf("+407%02d%06d", sub, tail)
	case "+44":
		// UK: +44 7xxx xxxxxx
		sub := rng.IntBetween(100, 999)
		tail := rng.IntBetween(100000, 999999)
		return fmt.Sprintf("+447%03d%06d", sub, tail)
	case "+49":
		// Germany: +49 15x xxxxxxx
		sub := rng.IntBetween(10, 79)
		tail := rng.IntBetween(1000000, 9999999)
		return fmt.Sprintf("+4915%02d%07d", sub, tail)
	default:
		return fmt.Sprintf("%s%09d", prefix, rng.IntBetween(100000000, 999999999))
	}
}
