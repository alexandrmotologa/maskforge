package stream

import (
	"strconv"
	"strings"
)

// EvaluateCondition evaluates a simple SQL-like boolean expression against a row's column values.
// Supports: =, !=, <>, >, <, >=, <=, LIKE, NOT LIKE, AND, OR.
// Example: "role != 'admin' AND is_active = 'true'"
func EvaluateCondition(expr string, row map[string]string) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true
	}

	// Handle OR expressions (lower precedence than AND)
	orClauses := splitTopLevel(expr, " OR ")
	if len(orClauses) > 1 {
		for _, orClause := range orClauses {
			if EvaluateCondition(orClause, row) {
				return true
			}
		}
		return false
	}

	// Handle AND expressions
	andClauses := splitTopLevel(expr, " AND ")
	if len(andClauses) > 1 {
		for _, andClause := range andClauses {
			if !EvaluateCondition(andClause, row) {
				return false
			}
		}
		return true
	}

	// Single predicate evaluation
	return evaluatePredicate(expr, row)
}

func evaluatePredicate(pred string, row map[string]string) bool {
	pred = strings.TrimSpace(pred)

	// List operators in order of length to avoid prefix collision (e.g. != before =)
	ops := []string{"!=", "<>", ">=", "<=", " NOT LIKE ", " LIKE ", "=", ">", "<"}

	for _, op := range ops {
		idx := strings.Index(strings.ToUpper(pred), strings.ToUpper(op))
		if idx != -1 {
			left := strings.TrimSpace(pred[:idx])
			right := strings.TrimSpace(pred[idx+len(op):])

			colName := strings.ToLower(strings.Trim(left, `"'` + "`"))
			rowVal, exists := row[colName]
			if !exists {
				// If column not found, try case-insensitive lookup
				for k, v := range row {
					if strings.EqualFold(k, colName) {
						rowVal = v
						exists = true
						break
					}
				}
			}

			// Clean right literal
			targetLiteral := strings.Trim(right, `"'` + "`")

			trimmedOp := strings.ToUpper(strings.TrimSpace(op))
			switch trimmedOp {
			case "=", "==":
				return strings.EqualFold(rowVal, targetLiteral)
			case "!=", "<>":
				return !strings.EqualFold(rowVal, targetLiteral)
			case "LIKE":
				return matchLikePattern(rowVal, targetLiteral)
			case "NOT LIKE":
				return !matchLikePattern(rowVal, targetLiteral)
			case ">":
				n1, e1 := strconv.ParseFloat(rowVal, 64)
				n2, e2 := strconv.ParseFloat(targetLiteral, 64)
				if e1 == nil && e2 == nil {
					return n1 > n2
				}
				return rowVal > targetLiteral
			case ">=":
				n1, e1 := strconv.ParseFloat(rowVal, 64)
				n2, e2 := strconv.ParseFloat(targetLiteral, 64)
				if e1 == nil && e2 == nil {
					return n1 >= n2
				}
				return rowVal >= targetLiteral
			case "<":
				n1, e1 := strconv.ParseFloat(rowVal, 64)
				n2, e2 := strconv.ParseFloat(targetLiteral, 64)
				if e1 == nil && e2 == nil {
					return n1 < n2
				}
				return rowVal < targetLiteral
			case "<=":
				n1, e1 := strconv.ParseFloat(rowVal, 64)
				n2, e2 := strconv.ParseFloat(targetLiteral, 64)
				if e1 == nil && e2 == nil {
					return n1 <= n2
				}
				return rowVal <= targetLiteral
			}
		}
	}

	return true
}

func matchLikePattern(str, pattern string) bool {
	str = strings.ToLower(str)
	pattern = strings.ToLower(pattern)

	if pattern == "%" {
		return true
	}
	if strings.HasPrefix(pattern, "%") && strings.HasSuffix(pattern, "%") {
		inner := pattern[1 : len(pattern)-1]
		return strings.Contains(str, inner)
	}
	if strings.HasPrefix(pattern, "%") {
		suffix := pattern[1:]
		return strings.HasSuffix(str, suffix)
	}
	if strings.HasSuffix(pattern, "%") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(str, prefix)
	}
	return str == pattern
}

func splitTopLevel(s, sep string) []string {
	var parts []string
	upper := strings.ToUpper(s)
	upperSep := strings.ToUpper(sep)

	start := 0
	for {
		idx := strings.Index(upper[start:], upperSep)
		if idx == -1 {
			parts = append(parts, strings.TrimSpace(s[start:]))
			break
		}
		parts = append(parts, strings.TrimSpace(s[start:start+idx]))
		start = start + idx + len(sep)
	}
	return parts
}
