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
