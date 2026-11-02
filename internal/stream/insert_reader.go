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
	Prefix     string // e.g. "INSERT INTO `users` (`id`, `email`) VALUES "
	Suffix     string // e.g. ";"
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

	// Table name can be in group 1 (backtick), 2 (quote), or 3 (identifier)
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

	// Determine the prefix before values section
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

	// Check if this table has any masking rules
	tc, found := cfg.GetTableConfig(parsed.TableName)
	if !found && cfg.Global.DefaultRule == "" {
		return line
	}

	// Map column index to rule
	rules := make(map[int]string)
	hasMasking := false

	if len(parsed.Columns) > 0 {
		for idx, col := range parsed.Columns {
			if rule, shouldMask := cfg.GetColumnRule(parsed.TableName, col); shouldMask {
				rules[idx] = rule
				hasMasking = true
			}
		}
	} else if len(tc.Columns) > 0 {
		// If columns are not explicitly listed in INSERT statement,
		// we match configured columns by index if possible
		return line
	}

	if !hasMasking {
		return line
	}

	// Extract tuples
	tuples := ExtractValuesTuples(parsed.ValuesPart)
	if len(tuples) == 0 {
		return line
	}

	var maskedTuples []string
	for _, tupleStr := range tuples {
		values := ParseSQLValuesTuple(tupleStr)
		if len(values) == 0 {
			maskedTuples = append(maskedTuples, tupleStr)
			continue
		}

		var formattedVals []string
		for colIdx, val := range values {
			ruleStr, shouldMask := rules[colIdx]
			if !shouldMask || val.IsNull {
				formattedVals = append(formattedVals, val.Raw)
				continue
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
