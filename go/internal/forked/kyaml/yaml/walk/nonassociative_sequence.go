// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"
)

// walkNonAssociativeSequence returns the value of VisitList
func (l Walker) walkNonAssociativeSequence() (*yaml.RNode, error) {
	return l.VisitList(l.Sources, l.Schema, NonAssociateList)
}
