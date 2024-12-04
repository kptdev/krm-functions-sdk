// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package merge2_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/kio/filters"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml/merge2"
)

var testCases = [][]testCase{scalarTestCases, listTestCases, elementTestCases, mapTestCases}

func TestMerge(t *testing.T) {
	for i := range testCases {
		for j := range testCases[i] {
			tc := testCases[i][j]
			t.Run(tc.description, func(t *testing.T) {
				actual, err := merge2.MergeStrings(tc.source, tc.dest, tc.infer, tc.mergeOptions)
				if !assert.NoError(t, err, tc.description) {
					t.FailNow()
				}
				e, err := filters.FormatInput(bytes.NewBufferString(tc.expected))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				estr := strings.TrimSpace(e.String())
				a, err := filters.FormatInput(bytes.NewBufferString(actual))
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				astr := strings.TrimSpace(a.String())
				if !assert.Equal(t, estr, astr, tc.description) {
					t.FailNow()
				}
			})
		}
	}
}

type testCase struct {
	description  string
	source       string
	dest         string
	expected     string
	infer        bool
	mergeOptions yaml.MergeOptions
}
