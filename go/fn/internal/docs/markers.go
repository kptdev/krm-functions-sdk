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

package docs

import "strings"

// Sections holds parsed README marker content.
type Sections struct {
	Short    string
	Long     string
	Examples string
}

const (
	markerShort    = "<!--mdtogo:Short-->"
	markerLong     = "<!--mdtogo:Long-->"
	markerExamples = "<!--mdtogo:Examples-->"
	markerEnd      = "<!--mdtogo-->"
)

// ParseMarkers extracts mdtogo marker sections from README content.
// Missing markers result in empty strings for the corresponding sections.
// When no markers are present at all, the full content (trimmed) is returned
// as the Long description.
func ParseMarkers(readme []byte) Sections {
	content := string(readme)

	short := extractSection(content, markerShort)
	long := extractSection(content, markerLong)
	examples := extractSection(content, markerExamples)

	// Fallback: if no markers are present at all, use full content as Long.
	if !hasAnyMarker(content) {
		return Sections{
			Long: strings.TrimSpace(content),
		}
	}

	return Sections{
		Short:    short,
		Long:     long,
		Examples: examples,
	}
}

// extractSection finds text between the given start marker and the next
// <!--mdtogo--> end marker. Returns empty string if either marker is missing.
func extractSection(content, startMarker string) string {
	startIdx := strings.Index(content, startMarker)
	if startIdx < 0 {
		return ""
	}
	afterStart := startIdx + len(startMarker)
	remaining := content[afterStart:]

	before, _, ok := strings.Cut(remaining, markerEnd)
	if !ok {
		return ""
	}

	return strings.TrimSpace(before)
}

// hasAnyMarker reports whether the content contains any mdtogo marker.
func hasAnyMarker(content string) bool {
	return strings.Contains(content, markerShort) ||
		strings.Contains(content, markerLong) ||
		strings.Contains(content, markerExamples) ||
		strings.Contains(content, markerEnd)
}
