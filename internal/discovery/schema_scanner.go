package discovery

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/maskforge/internal/config"
	"github.com/alexandrmotologa/maskforge/internal/faker"
	"github.com/alexandrmotologa/maskforge/internal/stream"
)

var (
	createTableRegex = regexp.MustCompile(`(?i)^CREATE\s+TABLE\s+(?:(?:` + "`" + `[^` + "`" + `]+` + "`" + `|"[^"]+"|[a-zA-Z0-9_]+)\.)?(?:` + "`" + `([^` + "`" + `]+)` + "`" + `|"([^"]+)"|([a-zA-Z0-9_]+))\s*\((.*)`)
)

// DiscoveredColumn describes a single detected table column.
type DiscoveredColumn struct {
	Name               string    `json:"name"`
	DetectedType       string    `json:"detected_type"`
	Risk               RiskLevel `json:"risk"`
	Compliance         []string  `json:"compliance"`
	RecommendedRule    string    `json:"recommended_rule"`
	Confidence         float64   `json:"confidence"`
	SampleValues       []string  `json:"sample_values"`
	MaskedSampleValues []string  `json:"masked_sample_values"`
}

// DiscoveredTable describes a database table scanned from SQL.
type DiscoveredTable struct {
	Name       string             `json:"name"`
	Columns    []DiscoveredColumn `json:"columns"`
	TotalRows  int64              `json:"total_rows"`
	SampleRows [][]string         `json:"sample_rows"`
}

// ScanReport represents the complete PII audit summary.
type ScanReport struct {
	Tables         []DiscoveredTable `json:"tables"`
	TotalTables    int               `json:"total_tables"`
	TotalColumns   int               `json:"total_columns"`
	PIIColumns     int               `json:"pii_columns"`
	CriticalCount  int               `json:"critical_count"`
	HighCount      int               `json:"high_count"`
	MediumCount    int               `json:"medium_count"`
	LowCount       int               `json:"low_count"`
	EstimatedRisk  RiskLevel         `json:"estimated_risk"`
	GenerationSalt string            `json:"generation_salt"`
}

// SchemaScanner extracts tables, columns, and sample values from a SQL stream.
type SchemaScanner struct {
	salt string
}

// NewSchemaScanner initializes a scanner.
func NewSchemaScanner(salt string) *SchemaScanner {
	if salt == "" {
		salt = "audit-default-salt"
	}
	return &SchemaScanner{salt: salt}
}

// ScanReader reads and analyzes SQL dump streams to discover PII.
func (s *SchemaScanner) ScanReader(r io.Reader, maxSamplesPerCol int) (*ScanReport, error) {
	if maxSamplesPerCol <= 0 {
		maxSamplesPerCol = 10
	}

	tablesMap := make(map[string]*discoveredTableBuilder)
	tableOrder := []string{}

	buf := bufio.NewReaderSize(r, 128*1024)
	reg := faker.NewRegistry()

	var activeCopyBlock *stream.PGCopyBlock
	var inCopyBlock bool

	for {
		line, isPrefix, err := buf.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		fullLine := line
		for isPrefix {
			var cont []byte
			cont, isPrefix, err = buf.ReadLine()
			if err != nil {
				return nil, err
			}
			fullLine = append(fullLine, cont...)
		}

		lineStr := string(fullLine)
		trimmed := strings.TrimSpace(lineStr)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		// Handle active COPY block
		if inCopyBlock {
			if bytes.Equal(bytes.TrimSpace(fullLine), []byte(`\.`)) {
				inCopyBlock = false
				activeCopyBlock = nil
				continue
			}

			tb, exists := tablesMap[activeCopyBlock.TableName]
			if exists {
				tb.totalRows++
				parts := bytes.Split(fullLine, []byte{'\t'})
				rowVals := make([]string, len(parts))
				for i, p := range parts {
					rowVals[i] = string(p)
				}

				if len(tb.sampleRows) < maxSamplesPerCol {
					tb.sampleRows = append(tb.sampleRows, rowVals)
				}

				for colIdx, colName := range activeCopyBlock.Columns {
					if colIdx < len(parts) {
						val := string(parts[colIdx])
						if len(tb.columnSamples[colName]) < maxSamplesPerCol {
							tb.columnSamples[colName] = append(tb.columnSamples[colName], val)
						}
					}
				}
			}
			continue
		}

		// Check COPY header
		if strings.HasPrefix(strings.ToUpper(trimmed), "COPY ") {
			dummyCfg := config.DefaultConfig()
			if block, ok := stream.ParsePGCopyHeader(lineStr, dummyCfg); ok {
				inCopyBlock = true
				activeCopyBlock = block

				tb, exists := tablesMap[block.TableName]
				if !exists {
					tb = newTableBuilder(block.TableName)
					tablesMap[block.TableName] = tb
					tableOrder = append(tableOrder, block.TableName)
				}
				for _, col := range block.Columns {
					tb.ensureColumn(col)
				}
				continue
			}
		}

		// Check INSERT statement
		if strings.HasPrefix(strings.ToUpper(trimmed), "INSERT INTO") {
			if parsed, ok := stream.ParseInsertStatement(lineStr); ok {
				tb, exists := tablesMap[parsed.TableName]
				if !exists {
					tb = newTableBuilder(parsed.TableName)
					tablesMap[parsed.TableName] = tb
					tableOrder = append(tableOrder, parsed.TableName)
				}
				for _, col := range parsed.Columns {
					tb.ensureColumn(col)
				}

				tuples := stream.ExtractValuesTuples(parsed.ValuesPart)
				for _, tuple := range tuples {
					tb.totalRows++
					values := stream.ParseSQLValuesTuple(tuple)
					rowVals := make([]string, len(values))
					for i, v := range values {
						rowVals[i] = v.Unquoted
						if i < len(parsed.Columns) {
							colName := parsed.Columns[i]
							if len(tb.columnSamples[colName]) < maxSamplesPerCol {
								tb.columnSamples[colName] = append(tb.columnSamples[colName], v.Unquoted)
							}
						}
					}
					if len(tb.sampleRows) < maxSamplesPerCol {
						tb.sampleRows = append(tb.sampleRows, rowVals)
					}
				}
			}
			continue
		}

		// Check CREATE TABLE statement
		if strings.HasPrefix(strings.ToUpper(trimmed), "CREATE TABLE") {
			matches := createTableRegex.FindStringSubmatch(trimmed)
			if len(matches) > 0 {
				tableName := matches[1]
				if tableName == "" {
					tableName = matches[2]
				}
				if tableName == "" {
					tableName = matches[3]
				}
				tableName = strings.Trim(tableName, `"'` + "`")

				tb, exists := tablesMap[tableName]
				if !exists {
					tb = newTableBuilder(tableName)
					tablesMap[tableName] = tb
					tableOrder = append(tableOrder, tableName)
				}
			}
		}
	}

	report := &ScanReport{
		Tables:         make([]DiscoveredTable, 0, len(tableOrder)),
		GenerationSalt: s.salt,
		EstimatedRisk:  RiskLow,
	}

	for _, tblName := range tableOrder {
		tb := tablesMap[tblName]
		dt := DiscoveredTable{
			Name:       tb.name,
			TotalRows:  tb.totalRows,
			SampleRows: tb.sampleRows,
			Columns:    make([]DiscoveredColumn, 0, len(tb.columnOrder)),
		}

		for _, colName := range tb.columnOrder {
			samples := tb.columnSamples[colName]
			piiRule, conf := DetectColumn(colName, samples)

			col := DiscoveredColumn{
				Name:         colName,
				SampleValues: samples,
			}

			if piiRule != nil {
				col.DetectedType = piiRule.Type
				col.Risk = piiRule.Risk
				col.Compliance = piiRule.Compliance
				col.RecommendedRule = piiRule.RecommendedRule
				col.Confidence = conf
				report.PIIColumns++

				switch piiRule.Risk {
				case RiskCritical:
					report.CriticalCount++
					report.EstimatedRisk = RiskCritical
				case RiskHigh:
					report.HighCount++
					if report.EstimatedRisk != RiskCritical {
						report.EstimatedRisk = RiskHigh
					}
				case RiskMedium:
					report.MediumCount++
					if report.EstimatedRisk != RiskCritical && report.EstimatedRisk != RiskHigh {
						report.EstimatedRisk = RiskMedium
					}
				case RiskLow:
					report.LowCount++
				}

				// Generate preview masked values
				for _, sample := range samples {
					ctx := faker.TransformContext{
						Salt:       s.salt,
						Table:      tb.name,
						Column:     colName,
						RawValue:   sample,
						RuleString: piiRule.RecommendedRule,
					}
					col.MaskedSampleValues = append(col.MaskedSampleValues, reg.Transform(ctx))
				}
			} else {
				col.DetectedType = "Non-Sensitive / Identifier"
				col.Risk = RiskLow
				col.RecommendedRule = "passthrough"
				col.MaskedSampleValues = samples
			}

			dt.Columns = append(dt.Columns, col)
			report.TotalColumns++
		}

		report.Tables = append(report.Tables, dt)
	}

	report.TotalTables = len(report.Tables)
	return report, nil
}

type discoveredTableBuilder struct {
	name          string
	columnOrder   []string
	columnSamples map[string][]string
	totalRows     int64
	sampleRows    [][]string
}

func newTableBuilder(name string) *discoveredTableBuilder {
	return &discoveredTableBuilder{
		name:          name,
		columnOrder:   []string{},
		columnSamples: make(map[string][]string),
	}
}

func (tb *discoveredTableBuilder) ensureColumn(col string) {
	col = strings.Trim(col, `"'` + "`")
	if _, exists := tb.columnSamples[col]; !exists {
		tb.columnSamples[col] = []string{}
		tb.columnOrder = append(tb.columnOrder, col)
	}
}
