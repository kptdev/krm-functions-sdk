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
	"reflect"
	"strings"
	"testing"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// kvPair is a single key/value entry in a generated ConfigMap's data map.
type kvPair struct {
	Key   string
	Value string
}

// genKVPair generates a YAML-safe key/value pair. Values are alphanumeric only
// so they cannot be misinterpreted by the YAML parser.
func genKVPair() gopter.Gen {
	return gen.Struct(reflect.TypeFor[kvPair](), map[string]gopter.Gen{
		"Key":   gen.RegexMatch(`[a-z][a-z0-9]{1,8}`),
		"Value": gen.RegexMatch(`[a-zA-Z0-9]{1,15}`),
	})
}

// genBoundedSliceOfN generates a slice of elements from elementGen whose length
// is between minLen and maxLen (inclusive). sliceType is the reflect type of the
// resulting slice (e.g. reflect.TypeOf([]kvPair{})).
func genBoundedSliceOfN(elementGen gopter.Gen, minLen, maxLen int, sliceType reflect.Type) gopter.Gen {
	return gen.IntRange(minLen, maxLen).FlatMap(func(n any) gopter.Gen {
		return gen.SliceOfN(n.(int), elementGen)
	}, sliceType)
}

// genKRMResource generates a valid KRM YAML resource (a ConfigMap) with random
// name and namespace. ConfigMaps are used because they are simple, always valid
// KRM resources that don't require complex schemas.
func genKRMResource() gopter.Gen {
	return gen.Struct(reflect.TypeFor[configMapParts](), map[string]gopter.Gen{
		"Name":      gen.RegexMatch(`[a-z][a-z0-9]{2,12}`),
		"Namespace": gen.RegexMatch(`[a-z][a-z0-9]{2,8}`),
		// Generate 1-3 data entries.
		"Entries": genBoundedSliceOfN(genKVPair(), 1, 3, reflect.TypeFor[[]kvPair]()),
	}).Map(func(parts configMapParts) string {
		var dataLines strings.Builder
		for _, e := range parts.Entries {
			fmt.Fprintf(&dataLines, "  %s: %s\n", e.Key, e.Value)
		}
		return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  namespace: %s
data:
%s`, parts.Name, parts.Namespace, dataLines.String())
	})
}

// configMapParts holds the generated pieces used to assemble a ConfigMap YAML.
type configMapParts struct {
	Name      string
	Namespace string
	Entries   []kvPair
}

// genKRMResourceList generates a slice of 1-5 valid KRM YAML resource strings.
func genKRMResourceList() gopter.Gen {
	return genBoundedSliceOfN(genKRMResource(), 1, 5, reflect.TypeFor[[]string]())
}

func TestProperty6_FileModeEquivalence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("file mode and STDIN mode produce equivalent output", prop.ForAll(
		func(resources []string) string {
			// --- File mode path ---
			// Write each resource to a temp file.
			tmpDir, err := os.MkdirTemp("", "property6-*")
			if err != nil {
				return fmt.Sprintf("failed to create temp dir: %v", err)
			}
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Errorf("failed to remove temp dir %s: %v", tmpDir, err)
				}
			}()

			var filePaths []string
			for i, res := range resources {
				path := filepath.Join(tmpDir, fmt.Sprintf("resource-%d.yaml", i))
				if err := os.WriteFile(path, []byte(res), 0600); err != nil {
					return fmt.Sprintf("failed to write temp file: %v", err)
				}
				filePaths = append(filePaths, path)
			}

			// Process via file mode: readFilesAsResourceList → Process → ToYAML.
			// Use a no-op processor that passes items through unchanged.
			noopProc := kubeobject.ResourceListProcessorFunc(func(rl *kubeobject.ResourceList) (bool, error) {
				return true, nil
			})

			fileRL, err := readFilesAsResourceList(filePaths)
			if err != nil {
				return fmt.Sprintf("readFilesAsResourceList failed: %v", err)
			}
			_, fnErr := noopProc.Process(fileRL)
			if fnErr != nil {
				return fmt.Sprintf("file mode Process failed: %v", fnErr)
			}
			fileOutput, err := fileRL.ToYAML()
			if err != nil {
				return fmt.Sprintf("file mode ToYAML failed: %v", err)
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
				return fmt.Sprintf("STDIN mode Run failed: %v\n  Input:\n%s", err, stdinInput.String())
			}

			// --- Compare outputs ---
			// Parse both outputs as ResourceLists and compare items.
			fileResultRL, err := kubeobject.ParseResourceList(fileOutput)
			if err != nil {
				return fmt.Sprintf("failed to parse file mode output: %v\n  Output:\n%s", err, string(fileOutput))
			}
			stdinResultRL, err := kubeobject.ParseResourceList(stdinOutput)
			if err != nil {
				return fmt.Sprintf("failed to parse STDIN mode output: %v\n  Output:\n%s", err, string(stdinOutput))
			}

			// Compare item counts.
			if len(fileResultRL.Items) != len(stdinResultRL.Items) {
				return fmt.Sprintf("item count mismatch: file mode has %d items, STDIN mode has %d items",
					len(fileResultRL.Items), len(stdinResultRL.Items))
			}

			// Compare each item by its string representation (after sorting, which
			// ToYAML does automatically).
			for i := range fileResultRL.Items {
				fileItem := fileResultRL.Items[i].String()
				stdinItem := stdinResultRL.Items[i].String()
				if fileItem != stdinItem {
					return fmt.Sprintf("item %d mismatch:\n  File mode:\n%s\n  STDIN mode:\n%s",
						i, fileItem, stdinItem)
				}
			}
			return ""
		},
		genKRMResourceList(),
	))

	properties.TestingRun(t)
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
