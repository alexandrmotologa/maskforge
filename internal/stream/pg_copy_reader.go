package stream

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/alexandrmotologa/maskforge/internal/config"
	"github.com/alexandrmotologa/maskforge/internal/faker"
)

var (
	// Matches: COPY [schema.]table [(col1, col2, ...)] FROM stdin;
	pgCopyRegex = regexp.MustCompile(`(?i)^COPY\s+(?:("[^"]+"|[a-zA-Z0-9_]+)\.)?("[^"]+"|[a-zA-Z0-9_]+)(?:\s*\(([^)]+)\))?\s+FROM\s+(?:stdin|STDIN)\s*;`)
)

// PGCopyBlock represents an active PostgreSQL COPY FROM stdin stream block.
type PGCopyBlock struct {
	TableName   string
	Columns     []string
	ColIndex    map[string]int
	Rules       map[int]string // col index -> rule string
	HasMasking  bool
}

// ParsePGCopyHeader tests whether a line is a PostgreSQL COPY ... FROM stdin statement.
func ParsePGCopyHeader(line string, cfg *config.Config) (*PGCopyBlock, bool) {
	trimmed := strings.TrimSpace(line)
	matches := pgCopyRegex.FindStringSubmatch(trimmed)
	if len(matches) == 0 {
		return nil, false
	}

	tableName := matches[2]
	tableName = strings.Trim(tableName, `"'`)

	var cols []string
	if matches[3] != "" {
		rawCols := strings.Split(matches[3], ",")
		for _, c := range rawCols {
			c = strings.TrimSpace(c)
			c = strings.Trim(c, `"'`)
			cols = append(cols, c)
		}
	}

	block := &PGCopyBlock{
		TableName: tableName,
		Columns:   cols,
		ColIndex:  make(map[string]int),
		Rules:     make(map[int]string),
	}

	for idx, col := range cols {
		block.ColIndex[col] = idx
		if rule, shouldMask := cfg.GetColumnRule(tableName, col); shouldMask {
			block.Rules[idx] = rule
			block.HasMasking = true
		}
	}

	return block, true
}

// ProcessPGCopyLine transforms a single tab-delimited COPY line.
func ProcessPGCopyLine(line []byte, block *PGCopyBlock, reg *faker.Registry, salt string) []byte {
	// If no columns need masking, return line as is
	if !block.HasMasking || len(block.Rules) == 0 {
		return line
	}

	// Split by tab delimiter
	parts := bytes.Split(line, []byte{'\t'})
	if len(parts) == 0 {
		return line
	}

	modified := false
	for colIdx, ruleStr := range block.Rules {
		if colIdx >= len(parts) {
			continue
		}

		rawVal := string(parts[colIdx])
		if rawVal == `\N` {
			// NULL value in COPY format
			continue
		}

		colName := ""
		if colIdx < len(block.Columns) {
			colName = block.Columns[colIdx]
		}

		ctx := faker.TransformContext{
			Salt:       salt,
			Table:      block.TableName,
			Column:     colName,
			RawValue:   rawVal,
			RuleString: ruleStr,
		}

		masked := reg.Transform(ctx)
		if masked != rawVal {
			parts[colIdx] = []byte(masked)
			modified = true
		}
	}

	if !modified {
		return line
	}

	return bytes.Join(parts, []byte{'\t'})
}

// StreamPGCopy reads and transforms lines in an active COPY block until the end-of-data marker "\." is reached.
func StreamPGCopy(reader *bufio.Reader, writer *bufio.Writer, block *PGCopyBlock, reg *faker.Registry, salt string) (int64, error) {
	var rowCount int64

	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return rowCount, err
		}

		// If line was truncated by ReadLine buffer, read the rest
		fullLine := line
		for isPrefix {
			var cont []byte
			cont, isPrefix, err = reader.ReadLine()
			if err != nil {
				return rowCount, err
			}
			fullLine = append(fullLine, cont...)
		}

		// Check for end of COPY data: line with exactly "\."
		if bytes.Equal(bytes.TrimSpace(fullLine), []byte(`\.`)) {
			writer.Write(fullLine)
			writer.WriteByte('\n')
			break
		}

		transformed := ProcessPGCopyLine(fullLine, block, reg, salt)
		writer.Write(transformed)
		writer.WriteByte('\n')
		rowCount++
	}

	return rowCount, nil
}
