// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package merge contains libraries for merging fields from one RNode to another
// RNode
package merge3

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml/walk"
)

func Merge(dest, original, update *yaml.RNode) (*yaml.RNode, error) {
	// if update == nil && original != nil => declarative deletion

	return walk.Walker{
		Visitor:            Visitor{},
		VisitKeysAsScalars: true,
		Sources:            []*yaml.RNode{dest, original, update}}.Walk()
}

func MergeStrings(dest, original, update string, infer bool) (string, error) {
	srcOriginal, err := yaml.Parse(original)
	if err != nil {
		return "", err
	}
	srcUpdated, err := yaml.Parse(update)
	if err != nil {
		return "", err
	}
	d, err := yaml.Parse(dest)
	if err != nil {
		return "", err
	}

	result, err := walk.Walker{
		InferAssociativeLists: infer,
		Visitor:               Visitor{},
		VisitKeysAsScalars:    true,
		Sources:               []*yaml.RNode{d, srcOriginal, srcUpdated}}.Walk()
	if err != nil {
		return "", err
	}
	return result.String()
}
