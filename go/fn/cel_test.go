// Copyright 2022 Google LLC
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveCEL(t *testing.T) {
	rl, err := ParseResourceList([]byte(`
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
    name: example2
    annotations:
      foo: bar
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: example2
    namespace: ns-1
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: example2
    namespace: ns-2
  data:
    mykey: myvalue
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: example3
    namespace: ns-2
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx-deployment
    labels:
      app: nginx
      hey: there
    annotations:
      config.kubernetes.io/index: '0'
      config.kubernetes.io/path: 'resources.yaml'
      internal.config.kubernetes.io/index: '0'
      internal.config.kubernetes.io/path: 'resources.yaml'
      internal.config.kubernetes.io/seqindent: 'compact'
  spec:
    replicas: 3
    selector:
      matchLabels:
        app: nginx
    paused: true
    strategy:
      type: Recreate
    template:
      metadata:
        labels:
          app: nginx
      spec:
        nodeSelector:
          disktype: ssd
        containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
          - containerPort: 80
    fakeStringSlice:
    - test1
    - test2
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testCases := map[string]struct {
		expr      string
		expResult string
		expError  string
	}{
		"string literal": {
			expr:      `"foo"`,
			expResult: "foo",
		},
		"int literal": {
			expr:     `15`,
			expError: "unsupported type conversion from 'int' to string",
		},
		"int literal conversion": {
			expr:      `string(15)`,
			expResult: "15",
		},
		"string var": {
			expr:      `items.filter(i, i.kind == "Deployment")[0].spec.strategy.type`,
			expResult: "Recreate",
		},
		"with_kind": {
			expr:      `items.with_kind("Deployment").spec.strategy.type`,
			expResult: "Recreate",
		},
		"with_name": {
			expr:      `items.with_name("nginx-deployment").spec.strategy.type`,
			expResult: "Recreate",
		},
		"with_namespace.with_name": {
			expr:      `items.with_namespace("ns-2").with_name("example2").data.mykey`,
			expResult: "myvalue",
		},
		"with_kind.with_name": {
			expr:      `items.with_kind("Namespace").with_name("example2").metadata.annotations["foo"]`,
			expResult: "bar",
		},
		"with_apiversion.with_name.with_kind": {
			expr:      `items.with_apiVersion("v1").with_name("example2").with_kind("Namespace").metadata.annotations["foo"]`,
			expResult: "bar",
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			result, err := rl.ResolveCEL(tc.expr)
			if tc.expError != "" {
				assert.EqualError(t, err, tc.expError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expResult, result)
			}
		})
	}
}
