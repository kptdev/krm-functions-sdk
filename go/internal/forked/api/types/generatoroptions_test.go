// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/types"
)

func TestMergeGlobalOptionsIntoLocal(t *testing.T) {
	tests := []struct {
		name     string
		local    *types.GeneratorOptions
		global   *types.GeneratorOptions
		expected *types.GeneratorOptions
	}{
		{
			name:     "everything nil",
			local:    nil,
			global:   nil,
			expected: nil,
		},
		{
			name: "nil global",
			local: &types.GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
			},
			global: nil,
			expected: &types.GeneratorOptions{
				Labels:                map[string]string{"pet": "dog"},
				Annotations:           map[string]string{"fruit": "apple"},
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
		},
		{
			name:  "nil local",
			local: nil,
			global: &types.GeneratorOptions{
				Labels:      map[string]string{"pet": "dog"},
				Annotations: map[string]string{"fruit": "apple"},
			},
			expected: &types.GeneratorOptions{
				Labels:                map[string]string{"pet": "dog"},
				Annotations:           map[string]string{"fruit": "apple"},
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
		},
		{
			name: "global doesn't damage local",
			local: &types.GeneratorOptions{
				Labels: map[string]string{"pet": "dog"},
				Annotations: map[string]string{
					"fruit": "apple"},
			},
			global: &types.GeneratorOptions{
				Labels: map[string]string{
					"pet":     "cat",
					"simpson": "homer",
				},
				Annotations: map[string]string{
					"fruit": "peach",
					"tesla": "Y",
				},
			},
			expected: &types.GeneratorOptions{
				Labels: map[string]string{
					"pet":     "dog",
					"simpson": "homer",
				},
				Annotations: map[string]string{
					"fruit": "apple",
					"tesla": "Y",
				},
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
		},
		{
			name: "global disable trumps local",
			local: &types.GeneratorOptions{
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
			global: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			expected: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name: "local disable works",
			local: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			global: &types.GeneratorOptions{
				DisableNameSuffixHash: false,
				Immutable:             false,
			},
			expected: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
		{
			name: "everyone wants disable",
			local: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			global: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
			expected: &types.GeneratorOptions{
				DisableNameSuffixHash: true,
				Immutable:             true,
			},
		},
	}
	for _, tc := range tests {
		actual := types.MergeGlobalOptionsIntoLocal(tc.local, tc.global)
		if !reflect.DeepEqual(tc.expected, actual) {
			t.Fatalf("%s annotations: Expected '%v', got '%v'",
				tc.name, tc.expected, *actual)
		}
	}
}
