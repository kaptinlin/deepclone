package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainPrintsCustomCloneableExample(t *testing.T) {
	// main redirects process stdout, so this test cannot run in parallel.
	output := captureOutput(t, main)

	assert.Contains(t, output, "=== Custom Cloneable Interface Example ===")
	assert.Contains(t, output, "Original: Value=10, Name=main")
	assert.Contains(t, output, "Cloned:   Value=11, Name=main_copy")
	assert.Contains(t, output, "Original: Name=Alice Modified, Age=30")
	assert.Contains(t, output, "Cloned:   Name=Alice, Age=30")
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
