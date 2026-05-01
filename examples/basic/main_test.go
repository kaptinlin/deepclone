package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPrintsBasicExamples(t *testing.T) {
	// main redirects process stdout, so this test cannot run in parallel.
	output := captureOutput(t, main)

	assert.Contains(t, output, "=== DeepClone Basic Examples ===")
	assert.Contains(t, output, "Original: 42, Cloned: 42")
	assert.Contains(t, output, "Original: [999 2 3 4 5]")
	assert.Contains(t, output, "Cloned:   [1 2 3 4 5]")
	assert.Contains(t, output, "Original: 200 (addr: ")
	assert.Contains(t, output, "Cloned:   100 (addr: ")
	assert.Contains(t, output, "Original: {Name:Jane Doe Age:30 Friends:[Charlie Bob]")
	assert.Contains(t, output, "Cloned:   {Name:John Doe Age:30 Friends:[Alice Bob]")
	assert.Contains(t, output, "Original: {Value:test Count:1}")
	assert.Contains(t, output, "Cloned:   {Value:test Count:2}")
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
