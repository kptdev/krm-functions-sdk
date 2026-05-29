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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// genNonEmptySectionContent generates arbitrary non-empty strings that do not
// contain mdtogo markers or newlines that would break containment checks.
func genNonEmptySectionContent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		s := rapid.StringMatching(`[a-zA-Z0-9 .,;:!?(){}\[\]'"/_-]{1,100}`).Draw(t, "content")
		// Ensure the generated content does not accidentally contain marker strings.
		s = strings.ReplaceAll(s, "<!--mdtogo:", "")
		s = strings.ReplaceAll(s, "<!--", "")
		// Ensure non-empty after trimming.
		s = strings.TrimSpace(s)
		if s == "" {
			s = "placeholder"
		}
		return s
	})
}

// genArbitrarySectionContent generates arbitrary strings (possibly empty) for
// use in property tests that don't require non-empty content.
func genArbitrarySectionContent() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		s := rapid.StringMatching(`[a-zA-Z0-9 .,;:!?(){}\[\]'"/_-]{0,100}`).Draw(t, "content")
		// Ensure the generated content does not accidentally contain the
		// forbidden strings we're testing for.
		s = strings.ReplaceAll(s, "Usage:", "")
		s = strings.ReplaceAll(s, "Flags:", "")
		return s
	})
}

func genArbitraryMetadata() *rapid.Generator[Metadata] {
	return rapid.Custom(func(t *rapid.T) Metadata {
		return Metadata{
			Image:       rapid.StringMatching(`[a-z0-9./:_-]{0,50}`).Draw(t, "image"),
			Description: rapid.StringMatching(`[a-zA-Z0-9 .,;:!?-]{0,80}`).Draw(t, "description"),
			Tags: rapid.SliceOfN(
				rapid.StringMatching(`[a-z0-9-]{1,20}`), 0, 5,
			).Draw(t, "tags"),
			SourceURL:          rapid.StringMatching(`https?://[a-z0-9./_ -]{0,50}`).Draw(t, "sourceURL"),
			ExamplePackageURLs: rapid.SliceOfN(rapid.StringMatching(`https?://[a-z0-9./_-]{0,50}`), 0, 3).Draw(t, "exampleURLs"),
			License:            rapid.StringMatching(`[A-Za-z0-9. -]{0,20}`).Draw(t, "license"),
			Hidden:             rapid.Bool().Draw(t, "hidden"),
		}
	})
}

func TestProperty4_HelpOutputExcludesCobraBoilerplate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate arbitrary sections (may be empty).
		sections := Sections{
			Short:    genArbitrarySectionContent().Draw(t, "short"),
			Long:     genArbitrarySectionContent().Draw(t, "long"),
			Examples: genArbitrarySectionContent().Draw(t, "examples"),
		}

		// Generate arbitrary metadata.
		meta := genArbitraryMetadata().Draw(t, "metadata")

		// Render help output.
		var buf bytes.Buffer
		RenderHelp(&buf, sections, meta)
		output := buf.String()

		// Assert that the help output does NOT contain cobra-style boilerplate.
		if strings.Contains(output, "Usage:") {
			t.Fatalf("help output contains forbidden 'Usage:' boilerplate\n  Output: %q", output)
		}
		if strings.Contains(output, "Flags:") {
			t.Fatalf("help output contains forbidden 'Flags:' boilerplate\n  Output: %q", output)
		}
	})
}

func TestProperty5_DocJSONContainsAllRequiredFields(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate non-empty section content for README markers.
		short := genNonEmptySectionContent().Draw(t, "short")
		long := genNonEmptySectionContent().Draw(t, "long")
		examples := genNonEmptySectionContent().Draw(t, "examples")

		// Format a README with valid mdtogo markers.
		readme := fmt.Sprintf(`<!--mdtogo:Short-->
%s
<!--mdtogo-->

<!--mdtogo:Long-->
%s
<!--mdtogo-->

<!--mdtogo:Examples-->
%s
<!--mdtogo-->
`, short, long, examples)

		// Parse the README to get sections.
		sections := ParseMarkers([]byte(readme))

		// Generate non-empty metadata fields to ensure all appear in output.
		meta := Metadata{
			Image:       rapid.StringMatching(`gcr\.io/[a-z0-9-]{3,20}/[a-z0-9-]{3,20}:v[0-9]+\.[0-9]+`).Draw(t, "image"),
			Description: genNonEmptySectionContent().Draw(t, "description"),
			Tags: rapid.SliceOfN(
				rapid.StringMatching(`[a-z]{3,10}`), 1, 5,
			).Draw(t, "tags"),
			SourceURL:          rapid.StringMatching(`https://github\.com/[a-z0-9-]{3,20}/[a-z0-9-]{3,20}`).Draw(t, "sourceURL"),
			ExamplePackageURLs: rapid.SliceOfN(rapid.StringMatching(`https://github\.com/[a-z0-9-]{3,20}/[a-z0-9-]{3,20}`), 1, 3).Draw(t, "exampleURLs"),
			License:            rapid.SampledFrom([]string{"Apache-2.0", "MIT", "BSD-3-Clause"}).Draw(t, "license"),
			Hidden:             rapid.Bool().Draw(t, "hidden"),
		}

		// Render doc JSON output.
		var buf bytes.Buffer
		err := RenderDoc(&buf, sections, meta)
		if err != nil {
			t.Fatalf("RenderDoc returned error: %v", err)
		}

		// Decode the JSON output.
		var output DocOutput
		if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
			t.Fatalf("failed to decode doc JSON: %v\n  Raw: %s", err, buf.String())
		}

		// Assert all non-empty source values from README appear in output.
		if output.Short != sections.Short {
			t.Fatalf("doc JSON 'short' mismatch\n  expected: %q\n  got: %q", sections.Short, output.Short)
		}
		if output.Long != sections.Long {
			t.Fatalf("doc JSON 'long' mismatch\n  expected: %q\n  got: %q", sections.Long, output.Long)
		}
		if output.Examples != sections.Examples {
			t.Fatalf("doc JSON 'examples' mismatch\n  expected: %q\n  got: %q", sections.Examples, output.Examples)
		}

		// Assert all non-empty source values from metadata appear in output.
		if output.Image != meta.Image {
			t.Fatalf("doc JSON 'image' mismatch\n  expected: %q\n  got: %q", meta.Image, output.Image)
		}
		if output.Description != meta.Description {
			t.Fatalf("doc JSON 'description' mismatch\n  expected: %q\n  got: %q", meta.Description, output.Description)
		}
		if len(output.Tags) != len(meta.Tags) {
			t.Fatalf("doc JSON 'tags' length mismatch\n  expected: %v\n  got: %v", meta.Tags, output.Tags)
		}
		for i, tag := range meta.Tags {
			if output.Tags[i] != tag {
				t.Fatalf("doc JSON 'tags[%d]' mismatch\n  expected: %q\n  got: %q", i, tag, output.Tags[i])
			}
		}
		if output.SourceURL != meta.SourceURL {
			t.Fatalf("doc JSON 'sourceURL' mismatch\n  expected: %q\n  got: %q", meta.SourceURL, output.SourceURL)
		}
		if len(output.ExamplePackageURLs) != len(meta.ExamplePackageURLs) {
			t.Fatalf("doc JSON 'examplePackageURLs' length mismatch\n  expected: %v\n  got: %v", meta.ExamplePackageURLs, output.ExamplePackageURLs)
		}
		for i, url := range meta.ExamplePackageURLs {
			if output.ExamplePackageURLs[i] != url {
				t.Fatalf("doc JSON 'examplePackageURLs[%d]' mismatch\n  expected: %q\n  got: %q", i, url, output.ExamplePackageURLs[i])
			}
		}
		if output.License != meta.License {
			t.Fatalf("doc JSON 'license' mismatch\n  expected: %q\n  got: %q", meta.License, output.License)
		}
		if output.Hidden != meta.Hidden {
			t.Fatalf("doc JSON 'hidden' mismatch\n  expected: %v\n  got: %v", meta.Hidden, output.Hidden)
		}
	})
}

func TestProperty3_HelpOutputContainsParsedSections(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		short := genNonEmptySectionContent().Draw(t, "short")
		long := genNonEmptySectionContent().Draw(t, "long")
		examples := genNonEmptySectionContent().Draw(t, "examples")

		// Format a README with valid mdtogo markers.
		readme := fmt.Sprintf(`<!--mdtogo:Short-->
%s
<!--mdtogo-->

<!--mdtogo:Long-->
%s
<!--mdtogo-->

<!--mdtogo:Examples-->
%s
<!--mdtogo-->
`, short, long, examples)

		// Parse the README to get sections (same as runtime would).
		sections := ParseMarkers([]byte(readme))

		// Render help output.
		var buf bytes.Buffer
		RenderHelp(&buf, sections, Metadata{})
		output := buf.String()

		// Assert that the help output contains each parsed section.
		if !strings.Contains(output, sections.Short) {
			t.Fatalf("help output does not contain Short section\n  Short: %q\n  Output: %q", sections.Short, output)
		}
		if !strings.Contains(output, sections.Long) {
			t.Fatalf("help output does not contain Long section\n  Long: %q\n  Output: %q", sections.Long, output)
		}
		if !strings.Contains(output, sections.Examples) {
			t.Fatalf("help output does not contain Examples section\n  Examples: %q\n  Output: %q", sections.Examples, output)
		}
	})
}

// --- Unit Tests for Renderers ---
// Validates: Requirements 2.3, 3.3, 3.6

func TestRenderHelp_FullSectionsAndMetadata(t *testing.T) {
	sections := Sections{
		Short:    "Set labels on all resources",
		Long:     "The set-labels function adds or updates labels on all resources in the package.",
		Examples: "  kpt fn eval --image gcr.io/kpt-fn/set-labels:v0.1 -- label_name=label_value",
	}
	meta := Metadata{
		Image:       "gcr.io/kpt-fn/set-labels:v0.1",
		Description: "Set labels on all resources",
		Tags:        []string{"mutator", "labels"},
	}

	var buf bytes.Buffer
	RenderHelp(&buf, sections, meta)
	output := buf.String()

	// Verify output contains the Short description.
	if !strings.Contains(output, sections.Short) {
		t.Errorf("expected output to contain Short %q, got:\n%s", sections.Short, output)
	}
	// Verify output contains the Long description.
	if !strings.Contains(output, sections.Long) {
		t.Errorf("expected output to contain Long %q, got:\n%s", sections.Long, output)
	}
	// Verify output contains the Examples content.
	if !strings.Contains(output, sections.Examples) {
		t.Errorf("expected output to contain Examples %q, got:\n%s", sections.Examples, output)
	}
	// Verify the "Examples:" header is present.
	if !strings.Contains(output, "Examples:") {
		t.Errorf("expected output to contain 'Examples:' header, got:\n%s", output)
	}
	// Verify no cobra boilerplate.
	if strings.Contains(output, "Usage:") {
		t.Errorf("output should not contain 'Usage:', got:\n%s", output)
	}
	if strings.Contains(output, "Flags:") {
		t.Errorf("output should not contain 'Flags:', got:\n%s", output)
	}
}

func TestRenderHelp_EmptySections(t *testing.T) {
	sections := Sections{}
	meta := Metadata{}

	var buf bytes.Buffer
	RenderHelp(&buf, sections, meta)
	output := buf.String()

	expected := "No documentation available. Pass fn.WithDocs to fn.AsMain to enable --help.\n"
	if output != expected {
		t.Errorf("expected minimal message %q, got %q", expected, output)
	}
}

func TestRenderDoc_ValidJSON_AllFields(t *testing.T) {
	sections := Sections{
		Short:    "Set labels on all resources",
		Long:     "The set-labels function adds or updates labels.",
		Examples: "  kpt fn eval --image gcr.io/kpt-fn/set-labels:v0.1",
	}
	meta := Metadata{
		Image:              "gcr.io/kpt-fn/set-labels:v0.1",
		Description:        "Set labels on all resources",
		Tags:               []string{"mutator", "labels"},
		SourceURL:          "https://github.com/kptdev/krm-functions/tree/main/functions/go/set-labels",
		ExamplePackageURLs: []string{"https://github.com/kptdev/krm-functions/tree/main/examples/set-labels-simple"},
		License:            "Apache-2.0",
		Hidden:             false,
	}

	var buf bytes.Buffer
	err := RenderDoc(&buf, sections, meta)
	if err != nil {
		t.Fatalf("RenderDoc returned error: %v", err)
	}

	// Verify output is valid JSON.
	var output DocOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("output is not valid JSON: %v\nRaw: %s", err, buf.String())
	}

	// Verify all fields are correctly populated.
	if output.Short != sections.Short {
		t.Errorf("Short: got %q, want %q", output.Short, sections.Short)
	}
	if output.Long != sections.Long {
		t.Errorf("Long: got %q, want %q", output.Long, sections.Long)
	}
	if output.Examples != sections.Examples {
		t.Errorf("Examples: got %q, want %q", output.Examples, sections.Examples)
	}
	if output.Image != meta.Image {
		t.Errorf("Image: got %q, want %q", output.Image, meta.Image)
	}
	if output.Description != meta.Description {
		t.Errorf("Description: got %q, want %q", output.Description, meta.Description)
	}
	if len(output.Tags) != len(meta.Tags) {
		t.Errorf("Tags length: got %d, want %d", len(output.Tags), len(meta.Tags))
	} else {
		for i, tag := range meta.Tags {
			if output.Tags[i] != tag {
				t.Errorf("Tags[%d]: got %q, want %q", i, output.Tags[i], tag)
			}
		}
	}
	if output.SourceURL != meta.SourceURL {
		t.Errorf("SourceURL: got %q, want %q", output.SourceURL, meta.SourceURL)
	}
	if len(output.ExamplePackageURLs) != len(meta.ExamplePackageURLs) {
		t.Errorf("ExamplePackageURLs length: got %d, want %d", len(output.ExamplePackageURLs), len(meta.ExamplePackageURLs))
	} else {
		for i, url := range meta.ExamplePackageURLs {
			if output.ExamplePackageURLs[i] != url {
				t.Errorf("ExamplePackageURLs[%d]: got %q, want %q", i, output.ExamplePackageURLs[i], url)
			}
		}
	}
	if output.License != meta.License {
		t.Errorf("License: got %q, want %q", output.License, meta.License)
	}
	if output.Hidden != meta.Hidden {
		t.Errorf("Hidden: got %v, want %v", output.Hidden, meta.Hidden)
	}
}

func TestRenderDoc_EmptyInputs(t *testing.T) {
	sections := Sections{}
	meta := Metadata{}

	var buf bytes.Buffer
	err := RenderDoc(&buf, sections, meta)
	if err != nil {
		t.Fatalf("RenderDoc returned error: %v", err)
	}

	// Verify output is valid JSON.
	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v\nRaw: %s", err, buf.String())
	}

	// Verify it decodes to a DocOutput with zero-value fields.
	var output DocOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("failed to decode into DocOutput: %v", err)
	}

	if output.Short != "" {
		t.Errorf("Short: got %q, want empty", output.Short)
	}
	if output.Long != "" {
		t.Errorf("Long: got %q, want empty", output.Long)
	}
	if output.Examples != "" {
		t.Errorf("Examples: got %q, want empty", output.Examples)
	}
	if output.Image != "" {
		t.Errorf("Image: got %q, want empty", output.Image)
	}
	if output.Hidden != false {
		t.Errorf("Hidden: got %v, want false", output.Hidden)
	}
}

func TestRenderDoc_HiddenFieldSerialization(t *testing.T) {
	sections := Sections{
		Short: "A hidden function",
	}
	meta := Metadata{
		Hidden: true,
	}

	var buf bytes.Buffer
	err := RenderDoc(&buf, sections, meta)
	if err != nil {
		t.Fatalf("RenderDoc returned error: %v", err)
	}

	// Verify the raw JSON contains "hidden": true.
	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v\nRaw: %s", err, buf.String())
	}

	hiddenVal, ok := raw["hidden"]
	if !ok {
		t.Fatal("JSON output does not contain 'hidden' field")
	}
	if hiddenVal != true {
		t.Errorf("hidden field: got %v, want true", hiddenVal)
	}

	// Also verify via struct decoding.
	var output DocOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("failed to decode into DocOutput: %v", err)
	}
	if !output.Hidden {
		t.Errorf("DocOutput.Hidden: got false, want true")
	}
}
