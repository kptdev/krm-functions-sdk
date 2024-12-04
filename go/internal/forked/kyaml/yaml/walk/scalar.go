// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import "github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"

// walkScalar returns the value of VisitScalar
func (l Walker) walkScalar() (*yaml.RNode, error) {
	return l.VisitScalar(l.Sources, l.Schema)
}
