package goatest

import (
	"bytes"
	"io"
	"sync"
	"time"
)

// threadSafeWriter is a thread-safe wrapper for io.Writer that captures output
type threadSafeWriter struct {
	mu       sync.RWMutex
	buffer   bytes.Buffer
	lines    []string
	delegate io.Writer // Optional delegate writer (like os.Stdout)
}

// newThreadSafeWriter creates a new thread-safe writer
func newThreadSafeWriter(delegate io.Writer) *threadSafeWriter {
	return &threadSafeWriter{
		delegate: delegate,
	}
}

// Write implements io.Writer interface
func (tsw *threadSafeWriter) Write(p []byte) (n int, err error) {
	tsw.mu.Lock()
	defer tsw.mu.Unlock()

	// Write to internal buffer
	n, err = tsw.buffer.Write(p)

	// Store individual lines (split by newlines to handle proper line boundaries)
	content := string(p)
	if content != "" {
		tsw.lines = append(tsw.lines, content)
	}

	// Write to delegate if provided (ignore errors to not affect internal buffering)
	if tsw.delegate != nil {
		_, _ = tsw.delegate.Write(p)
	}

	return n, err
}

// getOutput returns the complete captured output
func (tsw *threadSafeWriter) getOutput() string {
	tsw.mu.RLock()
	defer tsw.mu.RUnlock()
	return tsw.buffer.String()
}

// getLines returns all captured output lines
func (tsw *threadSafeWriter) getLines() []string {
	tsw.mu.RLock()
	defer tsw.mu.RUnlock()
	lines := make([]string, len(tsw.lines))
	copy(lines, tsw.lines)
	return lines
}

// containsOutput checks if the output contains the given string
func (tsw *threadSafeWriter) containsOutput(text string) bool {
	tsw.mu.RLock()
	defer tsw.mu.RUnlock()
	return bytes.Contains(tsw.buffer.Bytes(), []byte(text))
}

// waitForOutput waits for specific output to appear (with timeout)
func (tsw *threadSafeWriter) waitForOutput(text string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if tsw.containsOutput(text) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// reset clears the captured output
func (tsw *threadSafeWriter) reset() {
	tsw.mu.Lock()
	defer tsw.mu.Unlock()
	tsw.buffer.Reset()
	tsw.lines = tsw.lines[:0]
}
