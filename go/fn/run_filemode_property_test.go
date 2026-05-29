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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// genKRMResource generates a valid KRM YAML resource (a ConfigMap) with random
// name and namespace. ConfigMaps are used because they are simple, always valid
// KRM resources that don't require complex schemas.
func genKRMResource() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		name := rapid.StringMatching(`[a-z][a-z0-9]{2,12}`).Draw(t, "name")
		namespace := rapid.StringMatching(`[a-z][a-z0-9]{2,8}`).Draw(t, "namespace")
		// Generate 1-3 data entries with YAML-safe values (alphanumeric only,
		// no special characters that could be misinterpreted by the YAML parser).
		numEntries := rapid.IntRange(1, 3).Draw(t, "numEntries")
		var dataLines strings.Builder
		for i := range numEntries {
			key := rapid.StringMatching(`[a-z][a-z0-9]{1,8}`).Draw(t, fmt.Sprintf("key%d", i))
			value := rapid.StringMatching(`[a-zA-Z0-9]{1,15}`).Draw(t, fmt.Sprintf("value%d", i))
			dataLines.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
data:
%s`, name, namespace, dataLines.String())
	})
}

// genKRMResourceList generates a slice of 1-5 valid KRM YAML resource strings.
func genKRMResourceList() *rapid.Generator[[]string] {
	return rapid.SliceOfN(genKRMResource(), 1, 5)
}

func TestProperty6_FileModeEquivalence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		resources := genKRMResourceList().Draw(t, "resources")

		// --- File mode path ---
		// Write each resource to a temp file.
		tmpDir, err := os.MkdirTemp("", "property6-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		var filePaths []string
		for i, res := range resources {
			path := filepath.Join(tmpDir, fmt.Sprintf("resource-%d.yaml", i))
			if err := os.WriteFile(path, []byte(res), 0600); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}
			filePaths = append(filePaths, path)
		}

		// Process via file mode: readFilesAsResourceList → Process → ToYAML.
		// Use a no-op processor that passes items through unchanged.
		noopProc := ResourceListProcessorFunc(func(rl *ResourceList) (bool, error) {
			return true, nil
		})

		fileRL, err := readFilesAsResourceList(filePaths)
		if err != nil {
			t.Fatalf("readFilesAsResourceList failed: %v", err)
		}
		_, fnErr := noopProc.Process(fileRL)
		if fnErr != nil {
			t.Fatalf("file mode Process failed: %v", fnErr)
		}
		fileOutput, err := fileRL.ToYAML()
		if err != nil {
			t.Fatalf("file mode ToYAML failed: %v", err)
		}

		// --- STDIN mode path ---
		// Assemble the same resources into a ResourceList YAML (as STDIN would provide).
		var stdinInput strings.Builder
		stdinInput.WriteString("apiVersion: config.kubernetes.io/v1\nkind: ResourceList\nitems:\n")
		for _, res := range resources {
			// Indent each resource line under items as a YAML list element.
			stdinInput.WriteString("- ")
			first := true
			for _, line := range splitLines(res) {
				if first {
					stdinInput.WriteString(line + "\n")
					first = false
				} else {
					stdinInput.WriteString("  " + line + "\n")
				}
			}
		}

		stdinOutput, err := Run(noopProc, []byte(stdinInput.String()))
		if err != nil {
			t.Fatalf("STDIN mode Run failed: %v\n  Input:\n%s", err, stdinInput.String())
		}

		// --- Compare outputs ---
		// Parse both outputs as ResourceLists and compare items.
		fileResultRL, err := ParseResourceList(fileOutput)
		if err != nil {
			t.Fatalf("failed to parse file mode output: %v\n  Output:\n%s", err, string(fileOutput))
		}
		stdinResultRL, err := ParseResourceList(stdinOutput)
		if err != nil {
			t.Fatalf("failed to parse STDIN mode output: %v\n  Output:\n%s", err, string(stdinOutput))
		}

		// Compare item counts.
		if len(fileResultRL.Items) != len(stdinResultRL.Items) {
			t.Fatalf("item count mismatch: file mode has %d items, STDIN mode has %d items",
				len(fileResultRL.Items), len(stdinResultRL.Items))
		}

		// Compare each item by its string representation (after sorting, which
		// ToYAML does automatically).
		for i := range fileResultRL.Items {
			fileItem := fileResultRL.Items[i].String()
			stdinItem := stdinResultRL.Items[i].String()
			if fileItem != stdinItem {
				t.Fatalf("item %d mismatch:\n  File mode:\n%s\n  STDIN mode:\n%s",
					i, fileItem, stdinItem)
			}
		}
	})
}

// splitLines splits a string into lines, preserving empty lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
