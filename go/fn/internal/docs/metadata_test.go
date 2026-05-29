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
	"reflect"
	"testing"

	"go.yaml.in/yaml/v3"
	"pgregory.net/rapid"
)

// genSafeString generates simple alphanumeric strings suitable for YAML values
// that will not cause quoting or escaping issues.
func genSafeString() *rapid.Generator[string] {
	return rapid.StringMatching(`[a-zA-Z0-9][a-zA-Z0-9 ./_-]{0,50}`)
}

// genStringSlice generates a slice of safe strings (0–5 elements).
func genStringSlice() *rapid.Generator[[]string] {
	return rapid.SliceOfN(genSafeString(), 0, 5)
}

// genMetadata generates arbitrary valid Metadata structs.
func genMetadata() *rapid.Generator[Metadata] {
	return rapid.Custom(func(t *rapid.T) Metadata {
		return Metadata{
			Image:              genSafeString().Draw(t, "image"),
			Description:        genSafeString().Draw(t, "description"),
			Tags:               genStringSlice().Draw(t, "tags"),
			SourceURL:          genSafeString().Draw(t, "sourceURL"),
			ExamplePackageURLs: genStringSlice().Draw(t, "examplePackageURLs"),
			License:            genSafeString().Draw(t, "license"),
			Hidden:             rapid.Bool().Draw(t, "hidden"),
		}
	})
}

// --- Unit Tests for ParseMetadata ---

func TestParseMetadata_CompleteValid(t *testing.T) {
	input := []byte(`image: gcr.io/kpt-fn/set-labels:v0.1
description: Set labels on all resources
tags:
  - mutator
  - labels
sourceURL: https://github.com/kptdev/krm-functions/tree/main/functions/go/set-labels
examplePackageURLs:
  - https://github.com/kptdev/krm-functions/tree/main/examples/set-labels-simple
license: Apache-2.0
hidden: false
`)

	m, err := ParseMetadata(input)
	if err != nil {
		t.Fatalf("ParseMetadata returned unexpected error: %v", err)
	}

	if m.Image != "gcr.io/kpt-fn/set-labels:v0.1" {
		t.Errorf("Image = %q, want %q", m.Image, "gcr.io/kpt-fn/set-labels:v0.1")
	}
	if m.Description != "Set labels on all resources" {
		t.Errorf("Description = %q, want %q", m.Description, "Set labels on all resources")
	}
	if len(m.Tags) != 2 || m.Tags[0] != "mutator" || m.Tags[1] != "labels" {
		t.Errorf("Tags = %v, want [mutator labels]", m.Tags)
	}
	if m.SourceURL != "https://github.com/kptdev/krm-functions/tree/main/functions/go/set-labels" {
		t.Errorf("SourceURL = %q, want correct URL", m.SourceURL)
	}
	if len(m.ExamplePackageURLs) != 1 || m.ExamplePackageURLs[0] != "https://github.com/kptdev/krm-functions/tree/main/examples/set-labels-simple" {
		t.Errorf("ExamplePackageURLs = %v, want single URL", m.ExamplePackageURLs)
	}
	if m.License != "Apache-2.0" {
		t.Errorf("License = %q, want %q", m.License, "Apache-2.0")
	}
	if m.Hidden != false {
		t.Errorf("Hidden = %v, want false", m.Hidden)
	}
}

func TestParseMetadata_InvalidYAML(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "malformed YAML with bad indentation",
			input: []byte("tags:\n- valid\n bad: [unbalanced"),
		},
		{
			name:  "invalid YAML with tabs in the wrong places",
			input: []byte(":\n\t- :\n\t\t- [[["),
		},
		{
			name:  "scalar where mapping expected",
			input: []byte("image: !!binary \"not valid base64 %%%\""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := ParseMetadata(tc.input)
			if err == nil {
				t.Errorf("ParseMetadata did not return error for invalid YAML, got: %+v", m)
			}
			// Verify zero-value Metadata is returned on error
			if !reflect.DeepEqual(m, Metadata{}) {
				t.Errorf("ParseMetadata returned non-zero Metadata on error: %+v", m)
			}
		})
	}
}

func TestParseMetadata_PartialFields(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		validate func(t *testing.T, m Metadata)
	}{
		{
			name:  "only image field",
			input: []byte("image: gcr.io/kpt-fn/my-func:v1\n"),
			validate: func(t *testing.T, m Metadata) {
				if m.Image != "gcr.io/kpt-fn/my-func:v1" {
					t.Errorf("Image = %q, want %q", m.Image, "gcr.io/kpt-fn/my-func:v1")
				}
				if m.Description != "" {
					t.Errorf("Description = %q, want empty", m.Description)
				}
				if m.Tags != nil {
					t.Errorf("Tags = %v, want nil", m.Tags)
				}
				if m.License != "" {
					t.Errorf("License = %q, want empty", m.License)
				}
				if m.Hidden != false {
					t.Errorf("Hidden = %v, want false", m.Hidden)
				}
			},
		},
		{
			name:  "only tags and description",
			input: []byte("description: A function\ntags:\n  - validator\n"),
			validate: func(t *testing.T, m Metadata) {
				if m.Image != "" {
					t.Errorf("Image = %q, want empty", m.Image)
				}
				if m.Description != "A function" {
					t.Errorf("Description = %q, want %q", m.Description, "A function")
				}
				if len(m.Tags) != 1 || m.Tags[0] != "validator" {
					t.Errorf("Tags = %v, want [validator]", m.Tags)
				}
			},
		},
		{
			name:  "empty content",
			input: []byte(""),
			validate: func(t *testing.T, m Metadata) {
				if !reflect.DeepEqual(m, Metadata{}) {
					t.Errorf("expected zero-value Metadata for empty input, got: %+v", m)
				}
			},
		},
		{
			name:  "only license",
			input: []byte("license: MIT\n"),
			validate: func(t *testing.T, m Metadata) {
				if m.License != "MIT" {
					t.Errorf("License = %q, want %q", m.License, "MIT")
				}
				if m.Image != "" || m.Description != "" {
					t.Errorf("unexpected non-empty fields: Image=%q Description=%q", m.Image, m.Description)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := ParseMetadata(tc.input)
			if err != nil {
				t.Fatalf("ParseMetadata returned unexpected error: %v", err)
			}
			tc.validate(t, m)
		})
	}
}

func TestParseMetadata_HiddenFieldPropagation(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		hidden bool
	}{
		{
			name:   "hidden true",
			input:  []byte("image: gcr.io/kpt-fn/my-func:v1\nhidden: true\n"),
			hidden: true,
		},
		{
			name:   "hidden false explicit",
			input:  []byte("image: gcr.io/kpt-fn/my-func:v1\nhidden: false\n"),
			hidden: false,
		},
		{
			name:   "hidden absent defaults to false",
			input:  []byte("image: gcr.io/kpt-fn/my-func:v1\n"),
			hidden: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m, err := ParseMetadata(tc.input)
			if err != nil {
				t.Fatalf("ParseMetadata returned unexpected error: %v", err)
			}
			if m.Hidden != tc.hidden {
				t.Errorf("Hidden = %v, want %v", m.Hidden, tc.hidden)
			}
		})
	}
}

// --- Property-Based Tests ---

func TestProperty2_MetadataYAMLRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		original := genMetadata().Draw(t, "metadata")

		// Serialize to YAML
		data, err := yaml.Marshal(&original)
		if err != nil {
			t.Fatalf("failed to marshal Metadata to YAML: %v", err)
		}

		// Parse back with ParseMetadata
		parsed, err := ParseMetadata(data)
		if err != nil {
			t.Fatalf("ParseMetadata failed on valid YAML: %v\nYAML:\n%s", err, string(data))
		}

		// Assert equivalence for each field
		if parsed.Image != original.Image {
			t.Fatalf("Image mismatch:\n  got:  %q\n  want: %q", parsed.Image, original.Image)
		}
		if parsed.Description != original.Description {
			t.Fatalf("Description mismatch:\n  got:  %q\n  want: %q", parsed.Description, original.Description)
		}
		if parsed.SourceURL != original.SourceURL {
			t.Fatalf("SourceURL mismatch:\n  got:  %q\n  want: %q", parsed.SourceURL, original.SourceURL)
		}
		if parsed.License != original.License {
			t.Fatalf("License mismatch:\n  got:  %q\n  want: %q", parsed.License, original.License)
		}
		if parsed.Hidden != original.Hidden {
			t.Fatalf("Hidden mismatch:\n  got:  %v\n  want: %v", parsed.Hidden, original.Hidden)
		}

		// Compare Tags slices
		if len(parsed.Tags) != len(original.Tags) {
			t.Fatalf("Tags length mismatch:\n  got:  %v\n  want: %v", parsed.Tags, original.Tags)
		}
		for i := range original.Tags {
			if parsed.Tags[i] != original.Tags[i] {
				t.Fatalf("Tags[%d] mismatch:\n  got:  %q\n  want: %q", i, parsed.Tags[i], original.Tags[i])
			}
		}

		// Compare ExamplePackageURLs slices
		if len(parsed.ExamplePackageURLs) != len(original.ExamplePackageURLs) {
			t.Fatalf("ExamplePackageURLs length mismatch:\n  got:  %v\n  want: %v", parsed.ExamplePackageURLs, original.ExamplePackageURLs)
		}
		for i := range original.ExamplePackageURLs {
			if parsed.ExamplePackageURLs[i] != original.ExamplePackageURLs[i] {
				t.Fatalf("ExamplePackageURLs[%d] mismatch:\n  got:  %q\n  want: %q", i, parsed.ExamplePackageURLs[i], original.ExamplePackageURLs[i])
			}
		}
	})
}
