package faker

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

// CalculateLuhnCheckDigit calculates the checksum digit required to make a number Luhn-valid.
func CalculateLuhnCheckDigit(payload string) int {
	sum := 0
	double := true // Starting from rightmost payload digit, doubling every second digit

	for i := len(payload) - 1; i >= 0; i-- {
		d := int(payload[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}

	checksum := (10 - (sum % 10)) % 10
	return checksum
}

// IsValidLuhn checks whether a credit card number satisfies the Luhn mod-10 algorithm.
func IsValidLuhn(number string) bool {
	var digits []int
	for _, r := range number {
		if unicode.IsDigit(r) {
			digits = append(digits, int(r-'0'))
		}
	}
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}

	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

// GenerateCreditCard generates a deterministic credit card number compliant with Luhn mod-10.
// Supported brands: "visa", "mastercard", "amex", or "any".
func GenerateCreditCard(rng *DeterministicPRNG, brand string) string {
	b := strings.ToLower(brand)
	if b == "" || b == "any" {
		brands := []string{"visa", "mastercard"}
		b = rng.PickString(brands)
	}

	var prefix string
	var totalLen int

	switch b {
	case "visa":
		prefix = "4"
		totalLen = 16
	case "mastercard":
		prefixes := []string{"51", "52", "53", "54", "55"}
		prefix = rng.PickString(prefixes)
		totalLen = 16
	case "amex":
		prefixes := []string{"34", "37"}
		prefix = rng.PickString(prefixes)
		totalLen = 15
	default:
		prefix = "4"
		totalLen = 16
	}

	neededPayload := totalLen - 1 - len(prefix)
	payloadDigits := rng.Digits(neededPayload)
	fullPayload := prefix + payloadDigits

	checkDigit := CalculateLuhnCheckDigit(fullPayload)
	return fmt.Sprintf("%s%d", fullPayload, checkDigit)
}

// GenerateCVV generates a deterministic 3 or 4 digit card verification code.
func GenerateCVV(rng *DeterministicPRNG, length int) string {
	if length != 4 {
		length = 3
	}
	return rng.Digits(length)
}

// GenerateIBAN generates a valid IBAN with ISO 7064 Mod-97 check digits.
// Supported country codes: "RO", "DE", "FR", "GB", "NL", "ES", "IT" or default random.
func GenerateIBAN(rng *DeterministicPRNG, countryCode string) string {
	cc := strings.ToUpper(strings.TrimSpace(countryCode))
	if len(cc) != 2 {
		countries := []string{"RO", "DE", "FR", "GB", "NL", "ES"}
		cc = rng.PickString(countries)
	}

	var bban string
	switch cc {
	case "RO":
		// RO: 24 chars -> RO + 2 check + 4 bank + 16 account
		bankCodes := []string{"BTRL", "INGB", "RNCB", "BRDE", "RZBR"}
		bank := rng.PickString(bankCodes)
		account := rng.Digits(16)
		bban = bank + account
	case "DE":
		// DE: 22 chars -> DE + 2 check + 8 blz + 10 account
		blz := rng.Digits(8)
		account := rng.Digits(10)
		bban = blz + account
	case "FR":
		// FR: 27 chars -> FR + 2 check + 5 bank + 5 branch + 11 account + 2 rib
		bban = rng.Digits(23)
	case "GB":
		// GB: 22 chars -> GB + 2 check + 4 bank + 6 sort + 8 account
		bank := rng.PickString([]string{"BARC", "LOYD", "HBUK", "NWBK"})
		bban = bank + rng.Digits(14)
	default:
		bban = rng.Digits(16)
	}

	// Calculate Mod 97 check digits
	// Move country code and "00" to the end: BBAN + CC + 00
	rearranged := bban + cc + "00"
	numericStr := convertAlphaToDigits(rearranged)

	num := new(big.Int)
	num.SetString(numericStr, 10)
	mod := new(big.Int)
	mod.Mod(num, big.NewInt(97))

	checkNum := 98 - int(mod.Int64())
	checkDigits := fmt.Sprintf("%02d", checkNum)

	return cc + checkDigits + bban
}

func convertAlphaToDigits(input string) string {
	var sb strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) {
			val := int(unicode.ToUpper(r) - 'A' + 10)
			sb.WriteString(strconv.Itoa(val))
		} else if unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
