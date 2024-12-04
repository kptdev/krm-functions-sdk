// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/internal/utils"
)

func TestStringSliceIndex(t *testing.T) {
	assert.Equal(t, 0, utils.StringSliceIndex([]string{"a", "b"}, "a"))
	assert.Equal(t, 1, utils.StringSliceIndex([]string{"a", "b"}, "b"))
	assert.Equal(t, -1, utils.StringSliceIndex([]string{"a", "b"}, "c"))
	assert.Equal(t, -1, utils.StringSliceIndex([]string{}, "c"))
}

func TestStringSliceContains(t *testing.T) {
	assert.True(t, utils.StringSliceContains([]string{"a", "b"}, "a"))
	assert.True(t, utils.StringSliceContains([]string{"a", "b"}, "b"))
	assert.False(t, utils.StringSliceContains([]string{"a", "b"}, "c"))
	assert.False(t, utils.StringSliceContains([]string{}, "c"))
}

func TestSameEndingSubarray(t *testing.T) {
	assert.True(t, utils.SameEndingSubSlice([]string{"", "a", "b"}, []string{"a", "b"}))
	assert.True(t, utils.SameEndingSubSlice([]string{"a", "b", ""}, []string{"b", ""}))
	assert.True(t, utils.SameEndingSubSlice([]string{"a", "b"}, []string{"a", "b"}))
	assert.True(t, utils.SameEndingSubSlice([]string{"a", "b"}, []string{"b"}))
	assert.True(t, utils.SameEndingSubSlice([]string{"b"}, []string{"a", "b"}))
	assert.True(t, utils.SameEndingSubSlice([]string{}, []string{}))
	assert.False(t, utils.SameEndingSubSlice([]string{"a", "b"}, []string{"b", "a"}))
	assert.False(t, utils.SameEndingSubSlice([]string{"a", "b"}, []string{}))
	assert.False(t, utils.SameEndingSubSlice([]string{"a", "b"}, []string{""}))
}
