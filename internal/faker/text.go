package faker

import (
	"strconv"
	"strings"
)

var sampleWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit",
	"sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore",
	"magna", "aliqua", "enim", "ad", "minim", "veniam", "quis", "nostrud",
	"exercitation", "ullamco", "laboris", "nisi", "ut", "aliquip", "ex", "ea",
	"commodo", "consequat", "duis", "aute", "irure", "in", "reprehenderit",
}

// MaskPartial replaces the central characters of a string with a mask character (default '*').
func MaskPartial(rawValue string, prefixLen, suffixLen int, maskChar rune) string {
	runes := []rune(rawValue)
	length := len(runes)

	if length <= prefixLen+suffixLen {
		return strings.Repeat(string(maskChar), length)
	}

	var sb strings.Builder
	sb.WriteString(string(runes[:prefixLen]))
	for i := prefixLen; i < length-suffixLen; i++ {
		sb.WriteRune(maskChar)
	}
	sb.WriteString(string(runes[length-suffixLen:]))
	return sb.String()
}

// Redact returns a constant redacted label.
func Redact(customLabel string) string {
	if customLabel != "" {
		return customLabel
	}
	return "[REDACTED]"
}

// GenerateLoremText generates a deterministic phrase or sentence with given word count.
func GenerateLoremText(rng *DeterministicPRNG, wordsCount int) string {
	if wordsCount <= 0 {
		wordsCount = 8
	}
	words := make([]string, wordsCount)
	for i := 0; i < wordsCount; i++ {
		words[i] = rng.PickString(sampleWords)
	}
	if len(words) > 0 {
		words[0] = strings.Title(words[0])
	}
	return strings.Join(words, " ") + "."
}

// GenerateRandomNumberString generates a deterministic integer string in [min, max].
func GenerateRandomNumberString(rng *DeterministicPRNG, min, max int) string {
	val := rng.IntBetween(min, max)
	return strconv.Itoa(val)
}

// MaskDigits replaces all numeric digits with random digits, preserving formatting and non-digit characters.
func MaskDigits(rng *DeterministicPRNG, rawValue string) string {
	var sb strings.Builder
	for _, r := range rawValue {
		if r >= '0' && r <= '9' {
			sb.WriteString(strconv.Itoa(rng.Intn(10)))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// FormatPreservingString replaces alphabetic chars with deterministic random letters, keeping spaces and punctuation.
func FormatPreservingString(rng *DeterministicPRNG, rawValue string) string {
	var sb strings.Builder
	for _, r := range rawValue {
		switch {
		case r >= 'a' && r <= 'z':
			sb.WriteByte(byte('a' + rng.Intn(26)))
		case r >= 'A' && r <= 'Z':
			sb.WriteByte(byte('A' + rng.Intn(26)))
		case r >= '0' && r <= '9':
			sb.WriteByte(byte('0' + rng.Intn(10)))
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
