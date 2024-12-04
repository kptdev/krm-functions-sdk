// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package walk

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/openapi"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"
)

type ListKind int32

const (
	AssociativeList ListKind = 1 + iota
	NonAssociateList
)

// Visitor is invoked by walk with source and destination node pairs
type Visitor interface {
	VisitMap(Sources, *openapi.ResourceSchema) (*yaml.RNode, error)

	VisitScalar(Sources, *openapi.ResourceSchema) (*yaml.RNode, error)

	VisitList(Sources, *openapi.ResourceSchema, ListKind) (*yaml.RNode, error)
}

// ClearNode is returned if GrepFilter should do nothing after calling Set
var ClearNode *yaml.RNode
