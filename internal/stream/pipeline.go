package stream

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/alexandrmotologa/maskforge/internal/config"
	"github.com/alexandrmotologa/maskforge/internal/faker"
)

// PipelineMetrics tracks throughput and record counts during streaming.
type PipelineMetrics struct {
	LinesProcessed int64
	RowsMasked     int64
	BytesRead      int64
	BytesWritten   int64
	Duration       time.Duration
	RowsPerSec     float64
	MBPerSec       float64
}

// Pipeline coordinates stream parsing and anonymization.
type Pipeline struct {
	cfg      *config.Config
	registry *faker.Registry
}

// NewPipeline initializes a new streaming pipeline with a given configuration.
func NewPipeline(cfg *config.Config) *Pipeline {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &Pipeline{
		cfg:      cfg,
		registry: faker.NewRegistry(),
	}
}

// ProcessStream reads SQL lines from reader, masks sensitive columns, and writes to writer.
func (p *Pipeline) ProcessStream(r io.Reader, w io.Writer) (*PipelineMetrics, error) {
	startTime := time.Now()
	metrics := &PipelineMetrics{}

	bufReader := bufio.NewReaderSize(r, defaultBufferSize)
	countWriter := NewCountingWriter(w, defaultBufferSize)

	var inCopyBlock bool
	var activeCopyBlock *PGCopyBlock

	for {
		line, isPrefix, err := bufReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return metrics, err
		}

		// Read long lines fully
		fullLine := line
		for isPrefix {
			var cont []byte
			cont, isPrefix, err = bufReader.ReadLine()
			if err != nil {
				return metrics, err
			}
			fullLine = append(fullLine, cont...)
		}

		metrics.BytesRead += int64(len(fullLine) + 1)
		metrics.LinesProcessed++

		// Check if we are inside a COPY block
		if inCopyBlock {
			trimmed := bytes.TrimSpace(fullLine)
			if bytes.Equal(trimmed, []byte(`\.`)) {
				// End of COPY block
				inCopyBlock = false
				activeCopyBlock = nil
				countWriter.Write(fullLine)
				countWriter.WriteByte('\n')
				continue
			}

			// Process COPY row
			processed := ProcessPGCopyLine(fullLine, activeCopyBlock, p.registry, p.cfg.Salt)
			countWriter.Write(processed)
			countWriter.WriteByte('\n')
			metrics.RowsMasked++
			continue
		}

		// Check if line initiates a PostgreSQL COPY block
		lineStr := string(fullLine)
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(lineStr)), "COPY ") {
			if block, ok := ParsePGCopyHeader(lineStr, p.cfg); ok {
				inCopyBlock = true
				activeCopyBlock = block
				countWriter.Write(fullLine)
				countWriter.WriteByte('\n')
				continue
			}
		}

		// Check if line is an INSERT statement
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(lineStr)), "INSERT INTO") {
			processed := ProcessInsertStatement(lineStr, p.cfg, p.registry, p.cfg.Salt)
			countWriter.WriteString(processed)
			countWriter.WriteByte('\n')
			metrics.RowsMasked++
			continue
		}

		// Passthrough line (DDL, comments, settings, sequences)
		countWriter.Write(fullLine)
		countWriter.WriteByte('\n')
	}

	if err := countWriter.Flush(); err != nil {
		return metrics, err
	}

	metrics.Duration = time.Since(startTime)
	metrics.BytesWritten = countWriter.BytesWritten()

	seconds := metrics.Duration.Seconds()
	if seconds > 0 {
		metrics.RowsPerSec = float64(metrics.RowsMasked) / seconds
		metrics.MBPerSec = (float64(metrics.BytesRead) / (1024 * 1024)) / seconds
	}

	return metrics, nil
}
