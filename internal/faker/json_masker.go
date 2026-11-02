package faker

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MaskJSONPayload recursively masks targeted JSON paths within a JSON string.
// pathRules maps dot-delimited paths like "customer.email" or "data.user_id" to rule strings.
func MaskJSONPayload(rawJSON string, pathRules map[string]string, salt string, reg *Registry) (string, error) {
	trimmed := strings.TrimSpace(rawJSON)
	if trimmed == "" || trimmed == "null" || trimmed == "NULL" {
		return rawJSON, nil
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return rawJSON, fmt.Errorf("parsing JSON payload: %w", err)
	}

	for path, ruleStr := range pathRules {
		pathSegments := strings.Split(path, ".")
		applyJSONPathMask(parsed, pathSegments, ruleStr, salt, reg, path)
	}

	maskedBytes, err := json.Marshal(parsed)
	if err != nil {
		return rawJSON, fmt.Errorf("serializing masked JSON: %w", err)
	}

	return string(maskedBytes), nil
}

func applyJSONPathMask(node interface{}, segments []string, ruleStr, salt string, reg *Registry, fullPath string) {
	if len(segments) == 0 || node == nil {
		return
	}

	currKey := segments[0]

	switch val := node.(type) {
	case map[string]interface{}:
		child, exists := val[currKey]
		if !exists {
			return
		}

		if len(segments) == 1 {
			// Leaf node to transform
			strVal := fmt.Sprintf("%v", child)
			if child == nil {
				return
			}
			ctx := TransformContext{
				Salt:       salt,
				Table:      "json",
				Column:     fullPath,
				RawValue:   strVal,
				RuleString: ruleStr,
			}
			masked := reg.Transform(ctx)
			val[currKey] = masked
		} else {
			applyJSONPathMask(child, segments[1:], ruleStr, salt, reg, fullPath)
		}

	case []interface{}:
		// If current key is "*" or numeric index, or traverse all elements
		for _, item := range val {
			applyJSONPathMask(item, segments, ruleStr, salt, reg, fullPath)
		}
	}
}
