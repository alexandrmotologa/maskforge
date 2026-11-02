package stream

import (
	"regexp"
	"strings"

	"github.com/alexandrmotologa/maskforge/internal/config"
	"github.com/alexandrmotologa/maskforge/internal/faker"
)

var (
	// Matches: INSERT INTO [schema.]table [(cols)] VALUES ...
	insertHeaderRegex = regexp.MustCompile(`(?i)^INSERT\s+INTO\s+(?:(?:` + "`" + `[^` + "`" + `]+` + "`" + `|"[^"]+"|[a-zA-Z0-9_]+)\.)?(?:` + "`" + `([^` + "`" + `]+)` + "`" + `|"([^"]+)"|([a-zA-Z0-9_]+))(?:\s*\(([^)]+)\))?\s+VALUES\s*(.*)$`)
)

// ParsedInsert represents the components of a parsed SQL INSERT statement.
type ParsedInsert struct {
	Prefix     string
	Suffix     string
	TableName  string
	Columns    []string
	ValuesPart string
}

// ParseInsertStatement extracts table, columns, and values part from an INSERT query.
func ParseInsertStatement(line string) (*ParsedInsert, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(strings.ToUpper(trimmed), "INSERT INTO") {
		return nil, false
	}

	matches := insertHeaderRegex.FindStringSubmatch(trimmed)
	if len(matches) == 0 {
		return nil, false
	}

	tableName := matches[1]
	if tableName == "" {
		tableName = matches[2]
	}
	if tableName == "" {
		tableName = matches[3]
	}

	var cols []string
	if matches[4] != "" {
		rawCols := strings.Split(matches[4], ",")
		for _, c := range rawCols {
			c = strings.TrimSpace(c)
			c = strings.Trim(c, `"'` + "`")
			cols = append(cols, c)
		}
	}

	rawValuesPart := matches[5]
	suffix := ""
	if strings.HasSuffix(rawValuesPart, ";") {
		suffix = ";"
		rawValuesPart = strings.TrimSuffix(rawValuesPart, ";")
	}

	valuesIdx := strings.Index(strings.ToUpper(trimmed), "VALUES")
	prefix := trimmed[:valuesIdx+len("VALUES")] + " "

	return &ParsedInsert{
		Prefix:     prefix,
		Suffix:     suffix,
		TableName:  tableName,
		Columns:    cols,
		ValuesPart: strings.TrimSpace(rawValuesPart),
	}, true
}

// ProcessInsertStatement masks column values in an INSERT query according to the configuration.
func ProcessInsertStatement(line string, cfg *config.Config, reg *faker.Registry, salt string) string {
	parsed, ok := ParseInsertStatement(line)
	if !ok {
		return line
	}

	tc, found := cfg.GetTableConfig(parsed.TableName)
	if !found && cfg.Global.DefaultRule == "" {
		return line
	}

	rules := make(map[int]string)
	conditions := make(map[int]string)
	hasMasking := false

	if len(parsed.Columns) > 0 {
		for idx, col := range parsed.Columns {
			if rule, shouldMask := cfg.GetColumnRule(parsed.TableName, col); shouldMask {
				rules[idx] = rule
				hasMasking = true
				if cond := cfg.GetColumnCondition(parsed.TableName, col); cond != "" {
					conditions[idx] = cond
				}
			}
		}
	} else if len(tc.Columns) > 0 {
		return line
	}

	if !hasMasking {
		return line
	}

	tuples := ExtractValuesTuples(parsed.ValuesPart)
	if len(tuples) == 0 {
		return line
	}

	var maskedTuples []string
	for tupleIdx, tupleStr := range tuples {
		if tc.SampleLimit > 0 && int64(tupleIdx) >= tc.SampleLimit {
			break
		}

		values := ParseSQLValuesTuple(tupleStr)
		if len(values) == 0 {
			maskedTuples = append(maskedTuples, tupleStr)
			continue
		}

		// Build row lookup map for conditions
		var rowVals map[string]string
		if len(conditions) > 0 {
			rowVals = make(map[string]string, len(parsed.Columns))
			for i, col := range parsed.Columns {
				if i < len(values) {
					rowVals[strings.ToLower(col)] = values[i].Unquoted
				}
			}
		}

		var formattedVals []string
		for colIdx, val := range values {
			ruleStr, shouldMask := rules[colIdx]
			if !shouldMask || val.IsNull {
				formattedVals = append(formattedVals, val.Raw)
				continue
			}

			// Check condition
			if cond, hasCond := conditions[colIdx]; hasCond && cond != "" {
				if !EvaluateCondition(cond, rowVals) {
					formattedVals = append(formattedVals, val.Raw)
					continue
				}
			}

			colName := ""
			if colIdx < len(parsed.Columns) {
				colName = parsed.Columns[colIdx]
			}

			ctx := faker.TransformContext{
				Salt:       salt,
				Table:      parsed.TableName,
				Column:     colName,
				RawValue:   val.Unquoted,
				RuleString: ruleStr,
			}

			masked := reg.Transform(ctx)
			quoteChar := byte('\'')
			if len(val.Raw) > 0 && (val.Raw[0] == '"' || val.Raw[0] == '\'') {
				quoteChar = val.Raw[0]
			}
			formattedVals = append(formattedVals, FormatSQLValue(masked, val.IsString, val.IsNull, quoteChar))
		}

		maskedTuples = append(maskedTuples, "("+strings.Join(formattedVals, ", ")+")")
	}

	return parsed.Prefix + strings.Join(maskedTuples, ", ") + parsed.Suffix
}
