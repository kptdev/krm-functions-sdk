// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package patchstrategicmerge

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/kio"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml/merge2"
)

type Filter struct {
	Patch *yaml.RNode
}

var _ kio.Filter = Filter{}

// Filter does a strategic merge patch, which can delete nodes.
func (pf Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	var result []*yaml.RNode
	for i := range nodes {
		r, err := merge2.Merge(
			pf.Patch, nodes[i],
			yaml.MergeOptions{
				ListIncreaseDirection: yaml.MergeOptionsListPrepend,
			},
		)
		if err != nil {
			return nil, err
		}
		if r != nil {
			result = append(result, r)
		}
	}
	return result, nil
}
