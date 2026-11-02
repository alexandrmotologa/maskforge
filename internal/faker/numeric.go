package faker

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NumericJitter adds a deterministic +/- percent jitter to a numeric value while preserving decimal places.
func NumericJitter(rng *DeterministicPRNG, rawValue string, percent float64) string {
	if percent <= 0 {
		percent = 15.0
	}

	valStr := strings.TrimSpace(rawValue)
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return rawValue
	}

	// Calculate deterministic factor in [-percent, +percent]
	factor := (rng.Float64()*2.0 - 1.0) * (percent / 100.0)
	jittered := val * (1.0 + factor)

	// Detect number of decimal places in original
	decimals := 0
	if dotIdx := strings.IndexByte(valStr, '.'); dotIdx != -1 {
		decimals = len(valStr) - dotIdx - 1
	}

	if decimals > 0 {
		return fmt.Sprintf("%.*f", decimals, jittered)
	}
	return fmt.Sprintf("%.0f", math.Round(jittered))
}

// NumericRange maps a numeric value into a bounded [min, max] range deterministically.
func NumericRange(rng *DeterministicPRNG, minVal, maxVal float64, rawValue string) string {
	if maxVal <= minVal {
		maxVal = minVal + 100.0
	}
	val := minVal + rng.Float64()*(maxVal-minVal)

	decimals := 2
	valStr := strings.TrimSpace(rawValue)
	if dotIdx := strings.IndexByte(valStr, '.'); dotIdx != -1 {
		decimals = len(valStr) - dotIdx - 1
	}

	if decimals > 0 {
		return fmt.Sprintf("%.*f", decimals, val)
	}
	return fmt.Sprintf("%.0f", math.Round(val))
}
