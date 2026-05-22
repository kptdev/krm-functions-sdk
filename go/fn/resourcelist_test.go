// Copyright 2022 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fn

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

var dupResourceInput = []byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: Namespace
  metadata:
    name: example
- apiVersion: v1
  kind: Namespace
  metadata:
    name: example
`)

func TestCheckResourceDuplication(t *testing.T) {
	rl, _ := ParseResourceList(dupResourceInput)
	err := CheckResourceDuplication(rl)
	if err == nil {
		t.Errorf("expect to received duplicate error: got nil")
	}
	expectErr := "duplicate Resource(apiVersion=v1, kind=Namespace, Namespace=, Name=example)"
	if err.Error() != expectErr {
		t.Errorf("expect CheckResourceDuplication to fail; got %v, want %v", err, expectErr)
	}
}

func TestParseResourceListResults(t *testing.T) {
	rl, err := ParseResourceList([]byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: Namespace
  metadata:
    name: example
results:
- message: foo
  severity: error
- message: bar
  severity: warning
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := Results{
		{
			Message:  "foo",
			Severity: Error,
		},
		{
			Message:  "bar",
			Severity: Warning,
		},
	}
	if !cmp.Equal(expected, rl.Results) {
		t.Fatalf("unexpected diff: %v", cmp.Diff(expected, rl.Results))
	}
}

func TestParseZeroBytesResourceList(t *testing.T) {
	rl, err := ParseResourceList([]byte{})
	require.NoError(t, err)
	require.NotNil(t, rl)
}

// This caused a panic before
func TestEmptyToYAML(t *testing.T) {
	rl := &ResourceList{}

	var out []byte
	var err error
	require.NotPanics(t, func() {
		out, err = rl.ToYAML()
	})

	require.NoError(t, err)
	expected := fmt.Sprintf("apiVersion: %s\nkind: %s\n", kio.ResourceListAPIVersion, kio.ResourceListKind)
	require.Equal(t, expected, string(out))
}
