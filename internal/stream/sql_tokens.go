package stream

import (
	"strings"
	"unicode"
)

// SQLValue represents an extracted literal token from an INSERT VALUES tuple.
type SQLValue struct {
	Raw      string // e.g. "'john@example.com'", "42", "NULL"
	Unquoted string // e.g. "john@example.com"
	IsString bool
	IsNull   bool
}

// ParseSQLValuesTuple parses a single "(val1, val2, ...)" tuple into individual column values.
func ParseSQLValuesTuple(tupleStr string) []SQLValue {
	tupleStr = strings.TrimSpace(tupleStr)
	if strings.HasPrefix(tupleStr, "(") && strings.HasSuffix(tupleStr, ")") {
		tupleStr = tupleStr[1 : len(tupleStr)-1]
	}

	var values []SQLValue
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)
	escaped := false

	length := len(tupleStr)
	for i := 0; i < length; i++ {
		ch := tupleStr[i]

		if inQuote {
			current.WriteByte(ch)
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == quoteChar {
				// Check for SQL standard escaped quote: ''
				if i+1 < length && tupleStr[i+1] == quoteChar {
					current.WriteByte(tupleStr[i+1])
					i++ // Skip next quote
				} else {
					inQuote = false
				}
			}
		} else {
			if ch == '\'' || ch == '"' {
				inQuote = true
				quoteChar = ch
				current.WriteByte(ch)
			} else if ch == ',' {
				// End of value
				values = append(values, processRawSQLValue(current.String()))
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		}
	}

	if current.Len() > 0 || len(values) > 0 {
		values = append(values, processRawSQLValue(current.String()))
	}

	return values
}

func processRawSQLValue(raw string) SQLValue {
	trimmed := strings.TrimSpace(raw)
	upper := strings.ToUpper(trimmed)

	if upper == "NULL" || trimmed == "" {
		return SQLValue{
			Raw:      trimmed,
			Unquoted: "",
			IsString: false,
			IsNull:   true,
		}
	}

	if len(trimmed) >= 2 && ((trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') ||
		(trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"')) {
		unquoted := trimmed[1 : len(trimmed)-1]
		// Unescape standard SQL doubled quotes
		if trimmed[0] == '\'' {
			unquoted = strings.ReplaceAll(unquoted, "''", "'")
			unquoted = strings.ReplaceAll(unquoted, `\'`, `'`)
			unquoted = strings.ReplaceAll(unquoted, `\\`, `\`)
		}
		return SQLValue{
			Raw:      trimmed,
			Unquoted: unquoted,
			IsString: true,
			IsNull:   false,
		}
	}

	return SQLValue{
		Raw:      trimmed,
		Unquoted: trimmed,
		IsString: false,
		IsNull:   false,
	}
}

// FormatSQLValue formats an unquoted string back into a SQL-safe literal.
func FormatSQLValue(val string, wasString bool, wasNull bool, quoteChar byte) string {
	if wasNull {
		return "NULL"
	}
	if !wasString {
		// If it's a numeric or boolean literal, output directly
		if isNumericLiteral(val) || strings.EqualFold(val, "true") || strings.EqualFold(val, "false") {
			return val
		}
	}

	if quoteChar == 0 {
		quoteChar = '\''
	}

	// Escape single quotes for SQL: ' becomes ''
	escaped := strings.ReplaceAll(val, "'", "''")
	return string(quoteChar) + escaped + string(quoteChar)
}

func isNumericLiteral(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	dotSeen := false
	for i, r := range s {
		if i == 0 && (r == '-' || r == '+') {
			continue
		}
		if r == '.' {
			if dotSeen {
				return false
			}
			dotSeen = true
			continue
		}
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ExtractValuesTuples extracts each "(...)" group from the VALUES section of an INSERT statement.
func ExtractValuesTuples(valuesSection string) []string {
	var tuples []string
	var current strings.Builder
	depth := 0
	inQuote := false
	quoteChar := byte(0)
	escaped := false

	length := len(valuesSection)
	for i := 0; i < length; i++ {
		ch := valuesSection[i]

		if inQuote {
			current.WriteByte(ch)
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == quoteChar {
				if i+1 < length && valuesSection[i+1] == quoteChar {
					current.WriteByte(valuesSection[i+1])
					i++
				} else {
					inQuote = false
				}
			}
		} else {
			if ch == '\'' || ch == '"' {
				inQuote = true
				quoteChar = ch
				current.WriteByte(ch)
			} else if ch == '(' {
				depth++
				current.WriteByte(ch)
			} else if ch == ')' {
				depth--
				current.WriteByte(ch)
				if depth == 0 {
					tuples = append(tuples, current.String())
					current.Reset()
				}
			} else if depth > 0 {
				current.WriteByte(ch)
			}
		}
	}

	return tuples
}
