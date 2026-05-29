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

import (
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// For any three strings (short, long, examples), formatting them into a README
// with mdtogo markers and then parsing that README with ParseMarkers SHALL
// produce a Sections struct with fields equal to the original strings (after trimming).

// genSectionContent generates arbitrary non-empty strings that do not contain
// mdtogo markers (which would confuse the parser).
func genSectionContent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		s := rapid.StringMatching(`[a-zA-Z0-9 \t\n.,;:!?(){}\[\]'"/_-]{1,200}`).Draw(t, "content")
		// Ensure the generated content does not accidentally contain marker strings.
		s = strings.ReplaceAll(s, "<!--mdtogo:", "")
		s = strings.ReplaceAll(s, "<!--", "")
		return s
	})
}

// formatMarkedREADME formats three section strings into a README with mdtogo markers.
func formatMarkedREADME(short, long, examples string) string {
	return fmt.Sprintf(`<!--mdtogo:Short-->
%s
<!--mdtogo-->

<!--mdtogo:Long-->
%s
<!--mdtogo-->

<!--mdtogo:Examples-->
%s
<!--mdtogo-->
`, short, long, examples)
}

// genNoMarkerContent generates arbitrary strings guaranteed not to contain
// any mdtogo marker substrings.
func genNoMarkerContent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		s := rapid.StringMatching(`[a-zA-Z0-9 \t\n.,;:!?(){}\[\]'"/_-]{0,300}`).Draw(t, "content")
		// Strip anything that could form a marker.
		s = strings.ReplaceAll(s, "<!--mdtogo:", "")
		s = strings.ReplaceAll(s, "<!--mdtogo", "")
		s = strings.ReplaceAll(s, "<!--", "")
		return s
	})
}

func TestProperty7_MissingMarkersFallback(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		content := genNoMarkerContent().Draw(t, "readme")

		sections := ParseMarkers([]byte(content))

		if sections.Short != "" {
			t.Fatalf("Short should be empty for content without markers, got: %q", sections.Short)
		}
		if sections.Examples != "" {
			t.Fatalf("Examples should be empty for content without markers, got: %q", sections.Examples)
		}
		if got, want := sections.Long, strings.TrimSpace(content); got != want {
			t.Fatalf("Long mismatch:\n  got:  %q\n  want: %q", got, want)
		}
	})
}

func TestProperty1_MarkerParserRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		short := genSectionContent().Draw(t, "short")
		long := genSectionContent().Draw(t, "long")
		examples := genSectionContent().Draw(t, "examples")

		readme := formatMarkedREADME(short, long, examples)
		sections := ParseMarkers([]byte(readme))

		if got, want := sections.Short, strings.TrimSpace(short); got != want {
			t.Fatalf("Short mismatch:\n  got:  %q\n  want: %q", got, want)
		}
		if got, want := sections.Long, strings.TrimSpace(long); got != want {
			t.Fatalf("Long mismatch:\n  got:  %q\n  want: %q", got, want)
		}
		if got, want := sections.Examples, strings.TrimSpace(examples); got != want {
			t.Fatalf("Examples mismatch:\n  got:  %q\n  want: %q", got, want)
		}
	})
}
