package godock

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogCopier(t *testing.T) {
	t.Run("With separate stderr", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		copier := NewLogCopier(stdout, stderr)
		assert.NotNil(t, copier)
		assert.Equal(t, stdout, copier.stdout)
		assert.Equal(t, stderr, copier.stderr)
	})

	t.Run("With nil stderr", func(t *testing.T) {
		stdout := &bytes.Buffer{}
		copier := NewLogCopier(stdout, nil)
		assert.NotNil(t, copier)
		assert.Equal(t, stdout, copier.stdout)
		assert.Equal(t, stdout, copier.stderr)
	})
}

func createDockerLogEntry(streamType byte, content string) []byte {
	header := make([]byte, 8)
	header[0] = streamType
	binary.BigEndian.PutUint32(header[4:], uint32(len(content)))
	return append(header, []byte(content)...)
}

func TestLogCopier_Copy(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	copier := NewLogCopier(stdout, stderr)

	// Create a sample Docker log stream (multiplexed format)
	// Format: [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4} + content
	logStream := bytes.NewBuffer(nil)
	entry1 := createDockerLogEntry(1, "hello") // stdout
	entry2 := createDockerLogEntry(2, "world") // stderr
	logStream.Write(entry1)
	logStream.Write(entry2)

	// The written bytes should be the length of the original content
	// For entry1: len("hello") = 5
	// For entry2: len("world") = 5
	// Total: 10 bytes
	expectedBytes := int64(10)

	written, err := copier.Copy(logStream)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytes, written)
	assert.Equal(t, "hello", stdout.String())
	assert.Equal(t, "world", stderr.String())
}

func TestLogCopier_CopyWithPrefix(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	copier := NewLogCopier(stdout, stderr)

	// Create a sample Docker log stream (multiplexed format)
	logStream := bytes.NewBuffer(nil)
	entry1 := createDockerLogEntry(1, "hello") // stdout
	entry2 := createDockerLogEntry(2, "world") // stderr
	logStream.Write(entry1)
	logStream.Write(entry2)

	// The written bytes should be the length of the original content
	// Each entry has: content length only (not including prefixes)
	// For entry1: len("hello") = 5
	// For entry2: len("world") = 5
	// Total: 10 bytes
	expectedBytes := int64(10)

	written, err := copier.CopyWithPrefix(logStream, "[OUT] ", "[ERR] ")
	assert.NoError(t, err)
	assert.Equal(t, expectedBytes, written)
	assert.Equal(t, "[OUT] hello", stdout.String())
	assert.Equal(t, "[ERR] world", stderr.String())
}

func TestPrefixWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	writer := &prefixWriter{
		writer: &buf,
		prefix: "[TEST] ",
	}

	n, err := writer.Write([]byte("message"))
	assert.NoError(t, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, "[TEST] message", buf.String())

	// Test writing multiple lines
	buf.Reset()
	n, err = writer.Write([]byte("line1\nline2"))
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "[TEST] line1\nline2", buf.String())

	// Test error propagation
	errorWriter := &errorWriter{}
	writer.writer = errorWriter
	_, err = writer.Write([]byte("test"))
	assert.Error(t, err)
	assert.Equal(t, io.ErrShortWrite.Error(), err.Error())
}

// errorWriter is a helper type that always returns an error on Write
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}
