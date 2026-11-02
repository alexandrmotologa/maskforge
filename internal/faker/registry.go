package faker

import (
	"strconv"
	"strings"
)

// TransformContext encapsulates context needed to perform a deterministic masking operation.
type TransformContext struct {
	Salt       string
	Table      string
	Column     string
	RawValue   string
	RuleString string
}

// TransformerFunc is the signature of a transformation operation.
type TransformerFunc func(ctx TransformContext, params map[string]string) string

// Registry manages all registered masking transformers.
type Registry struct {
	transformers map[string]TransformerFunc
}

// NewRegistry initializes the default registry with all available transformers.
func NewRegistry() *Registry {
	r := &Registry{
		transformers: make(map[string]TransformerFunc),
	}
	r.registerDefaults()
	return r
}

func (r *Registry) Register(name string, fn TransformerFunc) {
	r.transformers[strings.ToLower(strings.TrimSpace(name))] = fn
}

func (r *Registry) registerDefaults() {
	// Passthrough
	r.Register("passthrough", func(ctx TransformContext, params map[string]string) string {
		return ctx.RawValue
	})
	r.Register("ignore", func(ctx TransformContext, params map[string]string) string {
		return ctx.RawValue
	})

	// Names
	r.Register("faker.first_name", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "first_name", ctx.RawValue)
		return GenerateFirstName(rng)
	})
	r.Register("faker.last_name", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "last_name", ctx.RawValue)
		return GenerateLastName(rng)
	})
	r.Register("faker.full_name", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "full_name", ctx.RawValue)
		return GenerateFullName(rng)
	})
	r.Register("faker.username", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "username", ctx.RawValue)
		return GenerateUsername(rng)
	})

	// Email
	r.Register("faker.email", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "email", ctx.RawValue)
		preserveDomain := params["preserve_domain"] == "true"
		customDomain := params["domain"]
		return GenerateEmail(rng, ctx.RawValue, preserveDomain, customDomain)
	})

	// Phone
	r.Register("faker.phone_number", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "phone", ctx.RawValue)
		prefix := params["prefix"]
		return GeneratePhoneNumber(rng, ctx.RawValue, prefix)
	})
	r.Register("faker.phone", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "phone", ctx.RawValue)
		return GeneratePhoneNumber(rng, ctx.RawValue, params["prefix"])
	})

	// Financial
	r.Register("faker.credit_card", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "credit_card", ctx.RawValue)
		brand := params["brand"]
		return GenerateCreditCard(rng, brand)
	})
	r.Register("faker.cvv", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "cvv", ctx.RawValue)
		length, _ := strconv.Atoi(params["length"])
		return GenerateCVV(rng, length)
	})
	r.Register("faker.iban", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "iban", ctx.RawValue)
		country := params["country"]
		return GenerateIBAN(rng, country)
	})

	// Identity
	r.Register("faker.ssn", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "ssn", ctx.RawValue)
		return GenerateSSN(rng)
	})
	r.Register("faker.cnp", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "cnp", ctx.RawValue)
		return GenerateCNP(rng)
	})
	r.Register("faker.national_id", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "national_id", ctx.RawValue)
		country := params["country"]
		return GenerateNationalID(rng, country)
	})

	// Address
	r.Register("faker.street_address", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "address", ctx.RawValue)
		return GenerateStreetAddress(rng)
	})
	r.Register("faker.address", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "address", ctx.RawValue)
		return GenerateStreetAddress(rng)
	})
	r.Register("faker.city", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "city", ctx.RawValue)
		return GenerateCity(rng)
	})
	r.Register("faker.country", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "country", ctx.RawValue)
		return GenerateCountry(rng)
	})
	r.Register("faker.zip_code", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "zip", ctx.RawValue)
		return GenerateZipCode(rng)
	})
	r.Register("faker.company", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "company", ctx.RawValue)
		return GenerateCompany(rng)
	})

	// Network
	r.Register("faker.ipv4", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "ipv4", ctx.RawValue)
		return GenerateIPv4(rng, params["mode"])
	})
	r.Register("faker.ipv6", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "ipv6", ctx.RawValue)
		return GenerateIPv6(rng)
	})
	r.Register("faker.mac", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "mac", ctx.RawValue)
		return GenerateMAC(rng)
	})
	r.Register("faker.uuid", func(ctx TransformContext, params map[string]string) string {
		return GenerateUUID(ctx.Salt, ctx.RawValue)
	})

	// Dates
	r.Register("faker.date_of_birth", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "dob", ctx.RawValue)
		minAge, _ := strconv.Atoi(params["min_age"])
		maxAge, _ := strconv.Atoi(params["max_age"])
		return GenerateDateOfBirth(rng, minAge, maxAge)
	})
	r.Register("date.jitter", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "jitter", ctx.RawValue)
		days, _ := strconv.Atoi(params["days"])
		return JitterDate(rng, ctx.RawValue, days)
	})

	// Hash & Password
	r.Register("hash.constant", func(ctx TransformContext, params map[string]string) string {
		return HashConstant(params["value"])
	})
	r.Register("hash.sha256", func(ctx TransformContext, params map[string]string) string {
		return HashSHA256(ctx.Salt, ctx.RawValue)
	})
	r.Register("hash.md5", func(ctx TransformContext, params map[string]string) string {
		return HashMD5(ctx.Salt, ctx.RawValue)
	})

	// Text & Redaction
	r.Register("mask.redact", func(ctx TransformContext, params map[string]string) string {
		return Redact(params["label"])
	})
	r.Register("mask.partial", func(ctx TransformContext, params map[string]string) string {
		prefixLen, _ := strconv.Atoi(params["prefix"])
		suffixLen, _ := strconv.Atoi(params["suffix"])
		if prefixLen == 0 && suffixLen == 0 {
			prefixLen = 2
			suffixLen = 2
		}
		maskChar := '*'
		if mStr := params["mask_char"]; len(mStr) > 0 {
			maskChar = []rune(mStr)[0]
		}
		return MaskPartial(ctx.RawValue, prefixLen, suffixLen, maskChar)
	})
	r.Register("mask.digits", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "digits", ctx.RawValue)
		return MaskDigits(rng, ctx.RawValue)
	})
	r.Register("faker.text", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "text", ctx.RawValue)
		words, _ := strconv.Atoi(params["words"])
		return GenerateLoremText(rng, words)
	})
	r.Register("random.number", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "rand_num", ctx.RawValue)
		minVal, _ := strconv.Atoi(params["min"])
		maxVal, _ := strconv.Atoi(params["max"])
		if maxVal <= minVal {
			maxVal = minVal + 100
		}
		return GenerateRandomNumberString(rng, minVal, maxVal)
	})

	// Numeric Jitter and Bounded Range
	r.Register("numeric.jitter", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "num_jitter", ctx.RawValue)
		pct, _ := strconv.ParseFloat(params["percent"], 64)
		if pct <= 0 {
			pct = 15.0
		}
		return NumericJitter(rng, ctx.RawValue, pct)
	})
	r.Register("numeric.range", func(ctx TransformContext, params map[string]string) string {
		rng := NewDeterministicPRNG(ctx.Salt, "num_range", ctx.RawValue)
		minVal, _ := strconv.ParseFloat(params["min"], 64)
		maxVal, _ := strconv.ParseFloat(params["max"], 64)
		return NumericRange(rng, minVal, maxVal, ctx.RawValue)
	})

	// Deep JSON / JSONB Masking
	r.Register("faker.json", func(ctx TransformContext, params map[string]string) string {
		if len(params) == 0 {
			return ctx.RawValue
		}
		masked, err := MaskJSONPayload(ctx.RawValue, params, ctx.Salt, r)
		if err != nil {
			return ctx.RawValue
		}
		return masked
	})
}

// ParseRule parses a rule string like "faker.email(preserve_domain=true, domain=test.com)"
// into the transformer name and a map of parameters.
func ParseRule(ruleStr string) (name string, params map[string]string) {
	ruleStr = strings.TrimSpace(ruleStr)
	params = make(map[string]string)

	openParen := strings.IndexByte(ruleStr, '(')
	closeParen := strings.LastIndexByte(ruleStr, ')')

	if openParen != -1 && closeParen != -1 && closeParen > openParen {
		name = strings.TrimSpace(ruleStr[:openParen])
		inner := ruleStr[openParen+1 : closeParen]
		// Parse key=value pairs or single positional argument
		parts := strings.Split(inner, ",")
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part, "=", 2)
				k := strings.ToLower(strings.TrimSpace(kv[0]))
				v := strings.Trim(strings.TrimSpace(kv[1]), `"'`)
				params[k] = v
			} else if part != "" {
				// Positional argument fallback
				if i == 0 {
					params["value"] = strings.Trim(part, `"'`)
				}
			}
		}
	} else {
		name = ruleStr
	}

	name = strings.ToLower(name)
	return name, params
}

// Transform applies the specified rule string to the given context.
func (r *Registry) Transform(ctx TransformContext) string {
	// Preserve NULL values
	if ctx.RawValue == `\N` || ctx.RawValue == "NULL" {
		return ctx.RawValue
	}

	name, params := ParseRule(ctx.RuleString)
	fn, ok := r.transformers[name]
	if !ok {
		// Fallback to passthrough if unknown rule
		return ctx.RawValue
	}

	return fn(ctx, params)
}
