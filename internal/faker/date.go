package faker

import (
	"fmt"
	"strings"
	"time"
)

// GenerateDateOfBirth generates a deterministic birthdate in YYYY-MM-DD format.
// Default age range is between minAge and maxAge (e.g. 18 and 68).
func GenerateDateOfBirth(rng *DeterministicPRNG, minAge, maxAge int) string {
	if minAge <= 0 {
		minAge = 18
	}
	if maxAge < minAge {
		maxAge = 70
	}
	baseYear := 2026
	birthYear := baseYear - rng.IntBetween(minAge, maxAge)
	month := rng.IntBetween(1, 12)
	maxDay := 28
	if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		maxDay = 31
	} else if month != 2 {
		maxDay = 30
	}
	day := rng.IntBetween(1, maxDay)
	return fmt.Sprintf("%04d-%02d-%02d", birthYear, month, day)
}

// JitterDate shifts an existing date by a deterministic +/- days offset to preserve general seasonality
// while anonymizing the exact timestamp.
func JitterDate(rng *DeterministicPRNG, originalValue string, maxDays int) string {
	if maxDays <= 0 {
		maxDays = 30
	}
	val := strings.TrimSpace(originalValue)
	val = strings.Trim(val, "'\"")

	var hasTime bool
	var layout string

	if len(val) >= 19 && strings.Contains(val, " ") {
		layout = "2006-01-02 15:04:05"
		hasTime = true
	} else if len(val) >= 10 {
		layout = "2006-01-02"
	} else {
		return GenerateDateOfBirth(rng, 18, 65)
	}

	t, err := time.Parse(layout, val[:len(layout)])
	if err != nil {
		return GenerateDateOfBirth(rng, 18, 65)
	}

	daysDelta := rng.IntBetween(-maxDays, maxDays)
	jittered := t.AddDate(0, 0, daysDelta)

	if hasTime {
		return jittered.Format("2006-01-02 15:04:05")
	}
	return jittered.Format("2006-01-02")
}
