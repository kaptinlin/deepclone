package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPrintsCircularReference(t *testing.T) {
	// main redirects process stdout, so this test cannot run in parallel.
	output := captureOutput(t, main)

	assert.Contains(t, output, "=== Circular Reference Example ===")
	assert.Contains(t, output, "Original: 1 -> 2 -> 1")
	assert.Contains(t, output, "Cloned: 1 -> 2 -> 1")
}

func captureOutput(t *testing.T, run func()) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
		_ = reader.Close()
		_ = writer.Close()
	}()

	run()
	require.NoError(t, writer.Close())

	var output bytes.Buffer
	_, err = io.Copy(&output, reader)
	require.NoError(t, err)

	return output.String()
}
