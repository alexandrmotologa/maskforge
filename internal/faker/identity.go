package faker

import (
	"fmt"
)

// GenerateSSN produces a validly formatted US Social Security Number (XXX-XX-XXXX)
// that adheres to official SSA exclusion rules.
func GenerateSSN(rng *DeterministicPRNG) string {
	// Area cannot be 000, 666, or in 900-999
	area := rng.IntBetween(1, 899)
	if area == 666 {
		area = 667
	}
	group := rng.IntBetween(1, 99)
	serial := rng.IntBetween(1, 9999)
	return fmt.Sprintf("%03d-%02d-%04d", area, group, serial)
}

// GenerateCNP produces a valid Romanian Personal Numeric Code (CNP) conforming to official checksum algorithm.
// S AA LL ZZ JJ NNN C
func GenerateCNP(rng *DeterministicPRNG) string {
	// S: 1/2 for 1900-1999, 5/6 for 2000-2099
	s := rng.PickString([]string{"1", "2", "5", "6"})
	year := rng.IntBetween(50, 99)
	if s == "5" || s == "6" {
		year = rng.IntBetween(0, 15)
	}
	month := rng.IntBetween(1, 12)
	day := rng.IntBetween(1, 28)
	county := rng.IntBetween(1, 46) // 01 to 46 (plus 51, 52 for Bucharest)
	order := rng.IntBetween(1, 999)

	prefix := fmt.Sprintf("%s%02d%02d%02d%02d%03d", s, year, month, day, county, order)

	// Calculate check digit with constant 279146358279
	weights := []int{2, 7, 9, 1, 4, 6, 3, 5, 8, 2, 7, 9}
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(prefix[i]-'0') * weights[i]
	}
	remainder := sum % 11
	checkDigit := remainder
	if remainder == 10 {
		checkDigit = 1
	}

	return fmt.Sprintf("%s%d", prefix, checkDigit)
}

// GenerateNationalID creates a generic national identifier or tax ID.
func GenerateNationalID(rng *DeterministicPRNG, country string) string {
	switch country {
	case "US":
		return GenerateSSN(rng)
	case "RO":
		return GenerateCNP(rng)
	default:
		return fmt.Sprintf("ID%s", rng.Digits(10))
	}
}
