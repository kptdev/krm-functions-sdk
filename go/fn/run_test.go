// Copyright 2026 The kpt Authors
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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestRunEmptyInputBytes(t *testing.T) {
	var noOpFn ResourceListProcessorFunc = func(rl *ResourceList) (bool, error) {
		return true, nil
	}

	output, err := Run(noOpFn, []byte{})
	require.NoError(t, err)
	expected := fmt.Appendf(nil, "apiVersion: %s\nkind: %s\n", kio.ResourceListAPIVersion, kio.ResourceListKind)
	assert.Equal(t, expected, output)
}

type myRLP struct{}

func (*myRLP) Process(*ResourceList) (bool, error) {
	return true, nil
}

func myRLPF(*ResourceList) (bool, error) {
	return true, nil
}

type myFR struct{}

func (*myFR) Run(*Context, *KubeObject, KubeObjects, *Results) bool {
	return true
}

var _ ResourceListProcessor = &myRLP{}
var _ ResourceListProcessorFunc = myRLPF
var _ Runner = &myFR{}

func TestAsMainTypes(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"cmd"}

	oldStdin := os.Stdin
	devNull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = devNull.Close()
	})
	os.Stdin = devNull

	testCases := map[string]any{
		"ResourceListProcessor":              &myRLP{},
		"Implicit ResourceListProcessorFunc": myRLPF,
		"Explicit ResourceListProcessorFunc": ResourceListProcessorFunc(myRLPF),
		"RunnerProcessor":                    runnerProcessor{ctx: t.Context(), fnRunner: &myFR{}},
		"Anonymous":                          func(*ResourceList) (bool, error) { return true, nil },
	}

	for name, input := range testCases {
		t.Run(name, func(t *testing.T) {
			err := AsMain(input)
			assert.NoError(t, err)
		})
	}
}
