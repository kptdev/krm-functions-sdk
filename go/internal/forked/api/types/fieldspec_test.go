// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/types"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/resid"
)

var mergeTests = []struct {
	name     string
	original types.FsSlice
	incoming types.FsSlice
	err      error
	result   types.FsSlice
}{
	{
		"normal",
		types.FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
		types.FsSlice{
			{
				Path:               "home",
				Gvk:                resid.Gvk{Group: "beans"},
				CreateIfNotPresent: false,
			},
		},
		nil,
		types.FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "home",
				Gvk:                resid.Gvk{Group: "beans"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		"ignore copy",
		types.FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
		types.FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
		},
		nil,
		types.FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
	},
	{
		"error on conflict",
		types.FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: false,
			},
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "pear"},
				CreateIfNotPresent: false,
			},
		},
		types.FsSlice{
			{
				Path:               "whatever",
				Gvk:                resid.Gvk{Group: "apple"},
				CreateIfNotPresent: true,
			},
		},
		fmt.Errorf("hey"),
		types.FsSlice{},
	},
}

func TestFsSlice_MergeAll(t *testing.T) {
	for _, item := range mergeTests {
		result, err := item.original.MergeAll(item.incoming)
		if item.err == nil {
			if err != nil {
				t.Fatalf("test %s: unexpected err %v", item.name, err)
			}
			if !reflect.DeepEqual(item.result, result) {
				t.Fatalf("test %s: expected: %v\n but got: %v\n",
					item.name, item.result, result)
			}
		} else {
			if err == nil {
				t.Fatalf("test %s: expected err: %v", item.name, err)
			}
			if !strings.Contains(err.Error(), "conflicting fieldspecs") {
				t.Fatalf("test %s: unexpected err: %v", item.name, err)
			}
		}
	}
}
