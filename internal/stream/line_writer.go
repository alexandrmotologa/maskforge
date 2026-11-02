package stream

import (
	"bufio"
	"io"
)

const defaultBufferSize = 256 * 1024 // 256 KB buffer for streaming

// CountingWriter wraps a bufio.Writer and counts bytes written.
type CountingWriter struct {
	writer *bufio.Writer
	count  int64
}

// NewCountingWriter creates a buffered counting writer.
func NewCountingWriter(w io.Writer, bufferSize int) *CountingWriter {
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	return &CountingWriter{
		writer: bufio.NewWriterSize(w, bufferSize),
	}
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.writer.Write(p)
	cw.count += int64(n)
	return n, err
}

func (cw *CountingWriter) WriteString(s string) (int, error) {
	n, err := cw.writer.WriteString(s)
	cw.count += int64(n)
	return n, err
}

func (cw *CountingWriter) WriteByte(c byte) error {
	err := cw.writer.WriteByte(c)
	if err == nil {
		cw.count++
	}
	return err
}

func (cw *CountingWriter) Flush() error {
	return cw.writer.Flush()
}

func (cw *CountingWriter) BytesWritten() int64 {
	return cw.count
}
