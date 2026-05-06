// Copyright 2025 The kpt Authors
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
	"encoding/json"
	"fmt"
	"io"
)

// DocOutput is the canonical JSON structure for --doc output.
// The SDK owns this schema as the single source of truth.
// Downstream consumers (e.g. catalog Hugo pipeline, metadata-schema.json)
// validate against a subset of these fields.
type DocOutput struct {
	Short              string   `json:"short"`
	Long               string   `json:"long"`
	Examples           string   `json:"examples"`
	Image              string   `json:"image"`
	Description        string   `json:"description"`
	Tags               []string `json:"tags"`
	SourceURL          string   `json:"sourceURL"`
	ExamplePackageURLs []string `json:"examplePackageURLs"`
	License            string   `json:"license"`
	Hidden             bool     `json:"hidden"`
}

// RenderHelp writes formatted help text to w.
// If sections are empty and metadata is zero-value, writes a minimal
// "no documentation available" message.
func RenderHelp(w io.Writer, sections Sections, meta Metadata) {
	if sections.Short == "" && sections.Long == "" && sections.Examples == "" && isMetadataEmpty(meta) {
		fmt.Fprint(w, "No documentation available. Pass fn.WithDocs to fn.AsMain to enable --help.\n")
		return
	}

	if sections.Short != "" {
		fmt.Fprintf(w, "%s\n", sections.Short)
	}

	if sections.Long != "" {
		if sections.Short != "" {
			fmt.Fprint(w, "\n")
		}
		fmt.Fprintf(w, "%s\n", sections.Long)
	}

	if sections.Examples != "" {
		if sections.Short != "" || sections.Long != "" {
			fmt.Fprint(w, "\n")
		}
		fmt.Fprintf(w, "Examples:\n%s\n", sections.Examples)
	}
}

// isMetadataEmpty reports whether all fields of meta are zero-value.
func isMetadataEmpty(meta Metadata) bool {
	return meta.Image == "" &&
		meta.Description == "" &&
		len(meta.Tags) == 0 &&
		meta.SourceURL == "" &&
		len(meta.ExamplePackageURLs) == 0 &&
		meta.License == "" &&
		!meta.Hidden
}

// RenderDoc writes JSON-encoded DocOutput to w.
// Returns error only if JSON encoding fails (shouldn't happen with these types).
func RenderDoc(w io.Writer, sections Sections, meta Metadata) error {
	out := DocOutput{
		Short:              sections.Short,
		Long:               sections.Long,
		Examples:           sections.Examples,
		Image:              meta.Image,
		Description:        meta.Description,
		Tags:               meta.Tags,
		SourceURL:          meta.SourceURL,
		ExamplePackageURLs: meta.ExamplePackageURLs,
		License:            meta.License,
		Hidden:             meta.Hidden,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}
