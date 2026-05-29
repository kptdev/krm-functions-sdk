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

import "go.yaml.in/yaml/v3"

// Metadata holds parsed metadata.yaml content.
type Metadata struct {
	Image              string   `yaml:"image" json:"image"`
	Description        string   `yaml:"description" json:"description"`
	Tags               []string `yaml:"tags" json:"tags"`
	SourceURL          string   `yaml:"sourceURL" json:"sourceURL"`
	ExamplePackageURLs []string `yaml:"examplePackageURLs" json:"examplePackageURLs"`
	License            string   `yaml:"license" json:"license"`
	Hidden             bool     `yaml:"hidden" json:"hidden"`
}

// ParseMetadata parses metadata.yaml content into a Metadata struct.
// Returns zero-value Metadata and an error if YAML is invalid.
// Returns successfully with partial fields if optional fields are missing.
func ParseMetadata(meta []byte) (Metadata, error) {
	var m Metadata
	if err := yaml.Unmarshal(meta, &m); err != nil {
		return Metadata{}, err
	}
	return m, nil
}
