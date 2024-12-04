package types_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/types"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/resid"
)

func TestPatchEquals(t *testing.T) {
	selector := types.Selector{
		ResId: resid.ResId{
			Gvk: resid.Gvk{
				Group:   "group",
				Version: "version",
				Kind:    "kind",
			},
			Name:      "name",
			Namespace: "namespace",
		},
		LabelSelector:      "selector",
		AnnotationSelector: "selector",
	}
	type testcase struct {
		patch1 types.Patch
		patch2 types.Patch
		expect bool
		name   string
	}
	testcases := []testcase{
		{
			name:   "empty patches",
			patch1: types.Patch{},
			patch2: types.Patch{},
			expect: true,
		},
		{
			name: "full patches",
			patch1: types.Patch{
				Path:  "foo",
				Patch: "bar",
				Target: &types.Selector{
					ResId: resid.ResId{
						Gvk: resid.Gvk{
							Group:   "group",
							Version: "version",
							Kind:    "kind",
						},
						Name:      "name",
						Namespace: "namespace",
					},
					LabelSelector:      "selector",
					AnnotationSelector: "selector",
				},
			},
			patch2: types.Patch{
				Path:  "foo",
				Patch: "bar",
				Target: &types.Selector{
					ResId: resid.ResId{
						Gvk: resid.Gvk{
							Group:   "group",
							Version: "version",
							Kind:    "kind",
						},
						Name:      "name",
						Namespace: "namespace",
					},
					LabelSelector:      "selector",
					AnnotationSelector: "selector",
				},
			},
			expect: true,
		},
		{
			name: "same target",
			patch1: types.Patch{
				Path:   "foo",
				Patch:  "bar",
				Target: &selector,
			},
			patch2: types.Patch{
				Path:   "foo",
				Patch:  "bar",
				Target: &selector,
			},
			expect: true,
		},
		{
			name: "omit target",
			patch1: types.Patch{
				Path:  "foo",
				Patch: "bar",
			},
			patch2: types.Patch{
				Path:  "foo",
				Patch: "bar",
			},
			expect: true,
		},
		{
			name: "one nil target",
			patch1: types.Patch{
				Path:   "foo",
				Patch:  "bar",
				Target: &selector,
			},
			patch2: types.Patch{
				Path:  "foo",
				Patch: "bar",
			},
			expect: false,
		},
		{
			name: "different path",
			patch1: types.Patch{
				Path: "foo",
			},
			patch2: types.Patch{
				Path: "bar",
			},
			expect: false,
		},
	}

	for _, tc := range testcases {
		if tc.expect != tc.patch1.Equals(tc.patch2) {
			t.Fatalf("%s: unexpected result %v", tc.name, !tc.expect)
		}
	}
}
