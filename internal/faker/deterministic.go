package faker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
)

// DeterministicPRNG is a seeded pseudo-random generator built on HMAC-SHA256.
// For any given salt and input raw value, it computes a stable stream of numbers.
type DeterministicPRNG struct {
	state uint64
	inc   uint64
}

// NewDeterministicPRNG initializes a PRNG from a salt and an input value.
// An optional context string (e.g. column name or domain) can be passed to
// ensure different columns deriving from the same raw value produce distinct outputs.
func NewDeterministicPRNG(salt, context, rawValue string) *DeterministicPRNG {
	h := hmac.New(sha256.New, []byte(salt))
	if context != "" {
		h.Write([]byte(context))
		h.Write([]byte{0})
	}
	h.Write([]byte(rawValue))
	sum := h.Sum(nil)

	// Derive two 64-bit words for a PCG-style deterministic generator
	seed := binary.BigEndian.Uint64(sum[0:8])
	inc := binary.BigEndian.Uint64(sum[8:16]) | 1 // must be odd

	rng := &DeterministicPRNG{
		state: 0,
		inc:   inc,
	}
	rng.step()
	rng.state += seed
	rng.step()

	return rng
}

func (p *DeterministicPRNG) step() {
	p.state = p.state*6364136223846793005 + p.inc
}

// Uint64 returns the next pseudo-random uint64.
func (p *DeterministicPRNG) Uint64() uint64 {
	oldstate := p.state
	p.step()
	// PCG XSH RR 64/32 adapted for 64-bit output
	xorshifted := ((oldstate >> 18) ^ oldstate) >> 27
	rot := oldstate >> 59
	v1 := uint32((xorshifted >> rot) | (xorshifted << ((-rot) & 31)))

	oldstate2 := p.state
	p.step()
	xorshifted2 := ((oldstate2 >> 18) ^ oldstate2) >> 27
	rot2 := oldstate2 >> 59
	v2 := uint32((xorshifted2 >> rot2) | (xorshifted2 << ((-rot2) & 31)))

	return (uint64(v1) << 32) | uint64(v2)
}

// Intn returns a deterministic integer in [0, n).
func (p *DeterministicPRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(p.Uint64() % uint64(n))
}

// IntBetween returns a deterministic integer in [min, max].
func (p *DeterministicPRNG) IntBetween(min, max int) int {
	if min >= max {
		return min
	}
	delta := max - min + 1
	return min + p.Intn(delta)
}

// Float64 returns a deterministic float64 in [0.0, 1.0).
func (p *DeterministicPRNG) Float64() float64 {
	return float64(p.Uint64()>>11) / (1 << 53)
}

// PickString selects one string from the provided slice deterministically.
func (p *DeterministicPRNG) PickString(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[p.Intn(len(items))]
}

// Digits generates a deterministic string of length n containing decimal digits.
func (p *DeterministicPRNG) Digits(length int) string {
	if length <= 0 {
		return ""
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = byte('0' + p.Intn(10))
	}
	return string(buf)
}
