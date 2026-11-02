package faker

import (
	"crypto/sha256"
	"fmt"
)

// GenerateIPv4 produces a safe synthetic IPv4 address (RFC 5737 documentation prefix or RFC 1918).
func GenerateIPv4(rng *DeterministicPRNG, mode string) string {
	switch mode {
	case "private":
		return fmt.Sprintf("10.%d.%d.%d", rng.IntBetween(0, 255), rng.IntBetween(0, 255), rng.IntBetween(1, 254))
	default:
		// RFC 5737 TEST-NET-1 (192.0.2.0/24) or TEST-NET-2 (198.51.100.0/24)
		if rng.Intn(2) == 0 {
			return fmt.Sprintf("192.0.2.%d", rng.IntBetween(1, 254))
		}
		return fmt.Sprintf("198.51.100.%d", rng.IntBetween(1, 254))
	}
}

// GenerateIPv6 produces a safe synthetic IPv6 address (RFC 3849 documentation prefix 2001:db8::).
func GenerateIPv6(rng *DeterministicPRNG) string {
	return fmt.Sprintf("2001:0db8:%04x:%04x:%04x:%04x:%04x:%04x",
		rng.IntBetween(0, 0xffff), rng.IntBetween(0, 0xffff),
		rng.IntBetween(0, 0xffff), rng.IntBetween(0, 0xffff),
		rng.IntBetween(0, 0xffff), rng.IntBetween(0, 0xffff))
}

// GenerateMAC produces a synthetic MAC address.
func GenerateMAC(rng *DeterministicPRNG) string {
	return fmt.Sprintf("02:%02x:%02x:%02x:%02x:%02x",
		rng.IntBetween(0, 255), rng.IntBetween(0, 255),
		rng.IntBetween(0, 255), rng.IntBetween(0, 255),
		rng.IntBetween(0, 255))
}

// GenerateUUID generates a deterministic valid UUIDv4 conforming to RFC 4122.
func GenerateUUID(salt, rawValue string) string {
	h := sha256.Sum256([]byte(salt + ":uuid:" + rawValue))
	b := h[:16]
	// Set version 4 (bits 4-7 of time_hi_and_version)
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant (bits 6-7 of clock_seq_hi_and_reserved)
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
