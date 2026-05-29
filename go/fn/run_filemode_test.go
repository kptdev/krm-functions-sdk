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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileMode_ValidInput verifies that valid input files produce correct output
// on STDOUT when processed via file mode.
// Requirements: 4.1, 4.2
func TestFileMode_ValidInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Write a valid ConfigMap resource to a temp file.
	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  key1: value1
`
	filePath := filepath.Join(tmpDir, "configmap.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(configMap), 0600))

	// Set os.Args to simulate file mode invocation.
	setArgs(t, []string{"cmd", filePath})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err, "file mode with valid input should succeed")
	})

	// Verify output is a valid ResourceList containing the ConfigMap.
	assert.NotEmpty(t, output, "file mode should produce output on STDOUT")
	assert.Contains(t, output, "kind: ResourceList")
	assert.Contains(t, output, "my-config")
	assert.Contains(t, output, "key1: value1")
}

// TestFileMode_MultipleFiles verifies that multiple valid input files are
// combined into a single ResourceList output.
// Requirements: 4.1, 4.2
func TestFileMode_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cm1 := `apiVersion: v1
kind: ConfigMap
metadata:
  name: config-one
  namespace: default
data:
  foo: bar
`
	cm2 := `apiVersion: v1
kind: ConfigMap
metadata:
  name: config-two
  namespace: default
data:
  baz: qux
`
	file1 := filepath.Join(tmpDir, "cm1.yaml")
	file2 := filepath.Join(tmpDir, "cm2.yaml")
	require.NoError(t, os.WriteFile(file1, []byte(cm1), 0600))
	require.NoError(t, os.WriteFile(file2, []byte(cm2), 0600))

	setArgs(t, []string{"cmd", file1, file2})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err, "file mode with multiple files should succeed")
	})

	// Both resources should appear in the output.
	assert.Contains(t, output, "config-one")
	assert.Contains(t, output, "config-two")
	assert.Contains(t, output, "foo: bar")
	assert.Contains(t, output, "baz: qux")
}

// TestFileMode_NonExistentFile verifies that a non-existent file returns a
// descriptive error message including the file path.
// Requirements: 4.3
func TestFileMode_NonExistentFile(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	setArgs(t, []string{"cmd", nonExistentPath})

	err := AsMain(noopProcessor)
	require.Error(t, err, "non-existent file should return an error")
	assert.Contains(t, err.Error(), "file not found")
	assert.Contains(t, err.Error(), nonExistentPath, "error should include the file path")
}

// TestFileMode_InvalidYAML verifies that a file with invalid YAML returns a
// parse error.
// Requirements: 4.3
func TestFileMode_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	invalidYAML := `{{{this is not valid YAML at all!!!`
	filePath := filepath.Join(tmpDir, "invalid.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(invalidYAML), 0600))

	setArgs(t, []string{"cmd", filePath})

	err := AsMain(noopProcessor)
	require.Error(t, err, "invalid YAML file should return an error")
	assert.Contains(t, err.Error(), filePath, "error should include the file path")
	assert.Contains(t, err.Error(), "failed to parse KRM resources from")
}

// TestFileMode_EmptyFile verifies that an empty file proceeds without error
// (valid for generators that don't require input items).
// Requirements: 4.1
func TestFileMode_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Write an empty file.
	filePath := filepath.Join(tmpDir, "empty.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(""), 0600))

	setArgs(t, []string{"cmd", filePath})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err, "empty file should proceed without error")
	})

	// Output should be a valid ResourceList (possibly with no items).
	assert.NotEmpty(t, output, "file mode should still produce output")
	assert.Contains(t, output, "kind: ResourceList")
}

// TestFileMode_WhitespaceOnlyFile verifies that a file containing only
// whitespace is treated as empty and proceeds without error.
// Requirements: 4.1
func TestFileMode_WhitespaceOnlyFile(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "whitespace.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte("   \n\n  \t  \n"), 0600))

	setArgs(t, []string{"cmd", filePath})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err, "whitespace-only file should proceed without error")
	})

	assert.Contains(t, output, "kind: ResourceList")
}

// TestFileMode_OutputToStdout verifies that file mode output goes to STDOUT
// (not STDERR or elsewhere).
// Requirements: 4.2
func TestFileMode_OutputToStdout(t *testing.T) {
	tmpDir := t.TempDir()

	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: stdout-test
  namespace: test
data:
  hello: world
`
	filePath := filepath.Join(tmpDir, "resource.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(configMap), 0600))

	setArgs(t, []string{"cmd", filePath})

	// Capture both stdout and stderr to verify output goes to stdout only.
	var stdoutOutput string
	stderrOutput := captureStderr(t, func() {
		stdoutOutput = captureStdout(t, func() {
			err := AsMain(noopProcessor)
			assert.NoError(t, err)
		})
	})

	// STDOUT should have the ResourceList output.
	assert.Contains(t, stdoutOutput, "stdout-test")
	assert.Contains(t, stdoutOutput, "kind: ResourceList")

	// STDERR should not contain the resource output.
	assert.NotContains(t, stderrOutput, "stdout-test")
}

// TestReadFilesAsResourceList_ValidFile tests the readFilesAsResourceList helper
// directly with a valid file.
func TestReadFilesAsResourceList_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()

	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: direct-test
  namespace: default
data:
  key: value
`
	filePath := filepath.Join(tmpDir, "cm.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(configMap), 0600))

	rl, err := readFilesAsResourceList([]string{filePath})
	require.NoError(t, err)
	require.NotNil(t, rl)

	// Should have one item.
	assert.Len(t, rl.Items, 1)
	assert.Equal(t, "direct-test", rl.Items[0].GetName())
	assert.Equal(t, "ConfigMap", rl.Items[0].GetKind())

	// FunctionConfig should be set (empty KubeObject).
	assert.NotNil(t, rl.FunctionConfig)
}

// TestReadFilesAsResourceList_NonExistentFile tests the readFilesAsResourceList
// helper directly with a non-existent file.
func TestReadFilesAsResourceList_NonExistentFile(t *testing.T) {
	nonExistentPath := "/tmp/definitely-does-not-exist-12345.yaml"

	rl, err := readFilesAsResourceList([]string{nonExistentPath})
	require.Error(t, err)
	assert.Nil(t, rl)
	assert.Contains(t, err.Error(), "file not found")
	assert.Contains(t, err.Error(), nonExistentPath)
}

// TestReadFilesAsResourceList_InvalidYAML tests the readFilesAsResourceList
// helper directly with invalid YAML content.
func TestReadFilesAsResourceList_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "bad.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(`{{{not yaml`), 0600))

	rl, err := readFilesAsResourceList([]string{filePath})
	require.Error(t, err)
	assert.Nil(t, rl)
	assert.Contains(t, err.Error(), "failed to parse KRM resources from")
	assert.Contains(t, err.Error(), filePath)
}

// TestReadFilesAsResourceList_EmptyFile tests the readFilesAsResourceList
// helper directly with an empty file.
func TestReadFilesAsResourceList_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "empty.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(""), 0600))

	rl, err := readFilesAsResourceList([]string{filePath})
	require.NoError(t, err)
	require.NotNil(t, rl)

	// Empty file should result in no items.
	assert.Empty(t, rl.Items)
	// FunctionConfig should still be set.
	assert.NotNil(t, rl.FunctionConfig)
}

// TestReadFilesAsResourceList_MultiDocument tests that a file with multiple
// YAML documents (separated by ---) produces multiple items.
func TestReadFilesAsResourceList_MultiDocument(t *testing.T) {
	tmpDir := t.TempDir()

	multiDoc := `apiVersion: v1
kind: ConfigMap
metadata:
  name: first
  namespace: default
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: second
  namespace: default
`
	filePath := filepath.Join(tmpDir, "multi.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(multiDoc), 0600))

	rl, err := readFilesAsResourceList([]string{filePath})
	require.NoError(t, err)
	require.NotNil(t, rl)

	assert.Len(t, rl.Items, 2)

	// Verify both items are present (order may vary).
	names := []string{rl.Items[0].GetName(), rl.Items[1].GetName()}
	assert.Contains(t, names, "first")
	assert.Contains(t, names, "second")
}

// TestFileMode_ProcessorReceivesItems verifies that the processor actually
// receives the items from the file and can modify them.
// Requirements: 4.1, 4.2
func TestFileMode_ProcessorReceivesItems(t *testing.T) {
	tmpDir := t.TempDir()

	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: to-be-labeled
  namespace: default
`
	filePath := filepath.Join(tmpDir, "cm.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte(configMap), 0600))

	// Use a processor that adds a label to all items.
	labelProc := ResourceListProcessorFunc(func(rl *ResourceList) (bool, error) {
		for _, item := range rl.Items {
			if err := item.SetLabel("added-by", "test"); err != nil {
				return false, err
			}
		}
		return true, nil
	})

	setArgs(t, []string{"cmd", filePath})

	output := captureStdout(t, func() {
		err := AsMain(labelProc)
		assert.NoError(t, err)
	})

	// Verify the label was added in the output.
	assert.Contains(t, output, "added-by")
	assert.Contains(t, output, "test")
}

// TestFileMode_HelpTakesPrecedence verifies that --help takes precedence over
// file paths when both are present.
func TestFileMode_HelpTakesPrecedence(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "cm.yaml")
	require.NoError(t, os.WriteFile(filePath, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n"), 0600))

	setArgs(t, []string{"cmd", "--help", filePath})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err)
	})

	// Should show help, not process the file.
	assert.Contains(t, output, "No documentation available")
	assert.NotContains(t, output, "kind: ResourceList")
}

// TestFileMode_MixedValidAndEmpty verifies that a mix of valid and empty files
// works correctly — only the valid file contributes items.
func TestFileMode_MixedValidAndEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: only-item
  namespace: default
`
	validFile := filepath.Join(tmpDir, "valid.yaml")
	emptyFile := filepath.Join(tmpDir, "empty.yaml")
	require.NoError(t, os.WriteFile(validFile, []byte(configMap), 0600))
	require.NoError(t, os.WriteFile(emptyFile, []byte(""), 0600))

	setArgs(t, []string{"cmd", emptyFile, validFile})

	output := captureStdout(t, func() {
		err := AsMain(noopProcessor)
		assert.NoError(t, err)
	})

	// Should contain the item from the valid file.
	assert.Contains(t, output, "only-item")
	// Should still be a valid ResourceList.
	assert.Contains(t, output, "kind: ResourceList")

	// Verify we can parse the output.
	rl, err := ParseResourceList([]byte(output))
	require.NoError(t, err)
	assert.Len(t, rl.Items, 1)
	assert.Equal(t, "only-item", rl.Items[0].GetName())
}

// TestFileMode_NonExistentAmongValid verifies that if one file in a list
// doesn't exist, the error is returned even if other files are valid.
// Requirements: 4.3
func TestFileMode_NonExistentAmongValid(t *testing.T) {
	tmpDir := t.TempDir()

	validFile := filepath.Join(tmpDir, "valid.yaml")
	require.NoError(t, os.WriteFile(validFile, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n"), 0600))

	nonExistent := filepath.Join(tmpDir, "missing.yaml")

	setArgs(t, []string{"cmd", validFile, nonExistent})

	// Capture stderr to suppress the error log from AsMain.
	captureStderr(t, func() {
		err := AsMain(noopProcessor)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
		assert.Contains(t, err.Error(), strings.TrimPrefix(nonExistent, ""))
	})
}
