// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fn

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kptdev/krm-functions-sdk/go/fn/internal/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noopProcessor is a minimal ResourceListProcessor that does nothing.
// Used to satisfy AsMain's input requirement without triggering STDIN reads.
var noopProcessor = ResourceListProcessorFunc(func(rl *ResourceList) (bool, error) {
	return true, nil
})

// captureStdout redirects os.Stdout to a pipe, runs fn, and returns what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	r.Close()

	return buf.String()
}

// captureStderr redirects os.Stderr to a pipe, runs fn, and returns what was written.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	r.Close()

	return buf.String()
}

// setArgs temporarily sets os.Args for the duration of a test.
func setArgs(t *testing.T, args []string) {
	t.Helper()
	origArgs := os.Args
	os.Args = args
	t.Cleanup(func() { os.Args = origArgs })
}

// TestAsMain_HelpFlag_ExitsZero verifies that --help returns nil (exit 0)
// without reading STDIN.
// Requirements: 2.1
func TestAsMain_HelpFlag_ExitsZero(t *testing.T) {
	setArgs(t, []string{"cmd", "--help"})

	// Close stdin to prove it is not read — if AsMain tries to read STDIN,
	// it would get an error or EOF immediately.
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	w.Close() // Close write end immediately — reading would get EOF
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close()
	})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor, WithDocs([]byte("some readme"), []byte("image: test")))
		assert.NoError(t, err, "--help should return nil (exit 0)")
	})

	// Should have produced some output
	assert.NotEmpty(t, output, "--help should produce output")
}

// TestAsMain_HelpFlag_NoDocs verifies that --help with no WithDocs prints
// a minimal message.
// Requirements: 2.3
func TestAsMain_HelpFlag_NoDocs(t *testing.T) {
	setArgs(t, []string{"cmd", "--help"})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err, "--help with no docs should return nil")
	})

	assert.Contains(t, output, "No documentation available")
	assert.Contains(t, output, "fn.WithDocs")
}

// TestAsMain_HelpFlag_WithDocs verifies that --help with WithDocs renders
// the README sections.
// Requirements: 2.2
func TestAsMain_HelpFlag_WithDocs(t *testing.T) {
	setArgs(t, []string{"cmd", "--help"})

	readme := []byte(`<!--mdtogo:Short-->
Set labels on resources
<!--mdtogo-->

<!--mdtogo:Long-->
The set-labels function adds labels to all resources.
<!--mdtogo-->

<!--mdtogo:Examples-->
  kpt fn eval --image set-labels:v0.1
<!--mdtogo-->
`)
	meta := []byte(`image: ghcr.io/kptdev/krm-functions-catalog/set-labels:v0.1
description: Set labels on all resources
`)

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor, WithDocs(readme, meta))
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Set labels on resources")
	assert.Contains(t, output, "set-labels function adds labels")
	assert.Contains(t, output, "kpt fn eval")
}

// TestAsMain_DocFlag_OutputsValidJSON verifies that --doc outputs valid JSON
// and returns nil (exit 0).
// Requirements: 3.1
func TestAsMain_DocFlag_OutputsValidJSON(t *testing.T) {
	setArgs(t, []string{"cmd", "--doc"})

	readme := []byte(`<!--mdtogo:Short-->
Set labels
<!--mdtogo-->

<!--mdtogo:Long-->
Long description here.
<!--mdtogo-->
`)
	meta := []byte(`image: ghcr.io/kptdev/krm-functions-catalog/set-labels:v0.1
description: Set labels on all resources
tags:
  - mutator
license: Apache-2.0
`)

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor, WithDocs(readme, meta))
		assert.NoError(t, err, "--doc should return nil (exit 0)")
	})

	// Verify that it is a valid JSON
	var docOutput docs.DocOutput
	err := json.Unmarshal([]byte(output), &docOutput)
	require.NoError(t, err, "--doc output should be valid JSON")

	// Verify fields are populated
	assert.Equal(t, "Set labels", docOutput.Short)
	assert.Equal(t, "Long description here.", docOutput.Long)
	assert.Equal(t, "ghcr.io/kptdev/krm-functions-catalog/set-labels:v0.1", docOutput.Image)
	assert.Equal(t, "Set labels on all resources", docOutput.Description)
	assert.Equal(t, []string{"mutator"}, docOutput.Tags)
	assert.Equal(t, "Apache-2.0", docOutput.License)
}

// TestAsMain_DocFlag_NoDocs verifies that --doc with no WithDocs outputs `{}`.
// Requirements: 3.3
func TestAsMain_DocFlag_NoDocs(t *testing.T) {
	setArgs(t, []string{"cmd", "--doc"})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err, "--doc with no docs should return nil")
	})

	assert.Equal(t, "{}", strings.TrimSpace(output))
}

// TestAsMain_DocFlag_HiddenField verifies that hidden:true in metadata
// propagates to the --doc JSON output.
// Requirements: 3.6
func TestAsMain_DocFlag_HiddenField(t *testing.T) {
	setArgs(t, []string{"cmd", "--doc"})

	readme := []byte(`<!--mdtogo:Short-->
Hidden function
<!--mdtogo-->
`)
	meta := []byte(`image: ghcr.io/kptdev/krm-functions-catalog/hidden-fn:v0.1
description: A hidden function
hidden: true
`)

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor, WithDocs(readme, meta))
		assert.NoError(t, err)
	})

	var docOutput docs.DocOutput
	err := json.Unmarshal([]byte(output), &docOutput)
	require.NoError(t, err, "--doc output should be valid JSON")

	assert.True(t, docOutput.Hidden, "hidden:true should propagate to JSON output")
	assert.Equal(t, "ghcr.io/kptdev/krm-functions-catalog/hidden-fn:v0.1", docOutput.Image)
}

// TestAsMain_DocFlag_InvalidMetadataYAML verifies that invalid metadata YAML
// logs a warning and continues with zero-value metadata (only README fields in output).
// Requirements: 5.5, 3.6
func TestAsMain_DocFlag_InvalidMetadataYAML(t *testing.T) {
	setArgs(t, []string{"cmd", "--doc"})

	readme := []byte(`<!--mdtogo:Short-->
My function
<!--mdtogo-->
`)
	invalidMeta := []byte(`{{{not valid yaml at all!!!`)

	var stdoutOutput string
	stderrOutput := captureStderr(t, func() {
		stdoutOutput = captureStdout(t, func() {
			err := AsMain(noopProcessor, WithDocs(readme, invalidMeta))
			assert.NoError(t, err, "invalid metadata should not cause AsMain to fail")
		})
	})

	// Verify warning was logged to stderr
	assert.Contains(t, stderrOutput, "warning")
	assert.Contains(t, stderrOutput, "invalid metadata YAML")

	// Verify JSON output still contains README fields
	var docOutput docs.DocOutput
	err := json.Unmarshal([]byte(stdoutOutput), &docOutput)
	require.NoError(t, err, "--doc output should still be valid JSON")

	assert.Equal(t, "My function", docOutput.Short)
	// Metadata fields should be zero-value
	assert.Empty(t, docOutput.Image)
	assert.Empty(t, docOutput.Description)
	assert.Empty(t, docOutput.Tags)
	assert.False(t, docOutput.Hidden)
}

// TestAsMain_HelpFlag_InvalidMetadataYAML verifies that --help with invalid
// metadata YAML logs a warning and continues rendering help from README only.
// Requirements: 5.5
func TestAsMain_HelpFlag_InvalidMetadataYAML(t *testing.T) {
	setArgs(t, []string{"cmd", "--help"})

	readme := []byte(`<!--mdtogo:Short-->
My function short desc
<!--mdtogo-->

<!--mdtogo:Long-->
Detailed description of the function.
<!--mdtogo-->
`)
	// Use YAML that actually fails to parse (unclosed flow mapping)
	invalidMeta := []byte(`{{{not valid yaml at all!!!`)

	var stdoutOutput string
	stderrOutput := captureStderr(t, func() {
		stdoutOutput = captureStdout(t, func() {
			err := AsMain(noopProcessor, WithDocs(readme, invalidMeta))
			assert.NoError(t, err, "invalid metadata should not cause --help to fail")
		})
	})

	// Verify warning was logged
	assert.Contains(t, stderrOutput, "warning")
	assert.Contains(t, stderrOutput, "invalid metadata YAML")

	// Verify help output still contains README sections
	assert.Contains(t, stdoutOutput, "My function short desc")
	assert.Contains(t, stdoutOutput, "Detailed description of the function")
}
