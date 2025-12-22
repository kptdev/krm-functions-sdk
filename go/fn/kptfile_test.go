// Copyright 2024 The kpt and Nephio Authors
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
	"strings"
	"testing"

	kptfileapi "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"gotest.tools/assert"
)

func TestAddCondition(t *testing.T) {
	testcases := []struct {
		name            string
		cond            kptfileapi.Condition
		resources       map[string]string
		expectedKptfile string
	}{
		{
			name: "add condition to missing status",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-labels:latest
      configPath: fn-config.yaml`,

				"service.yaml": `
apiVersion: v1
kind: Service
metadata:
 name: whatever
 labels:
   app: myApp`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
  annotations:
    config.kubernetes.io/local-config: "true"
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/set-labels:latest
      configPath: fn-config.yaml
status:
  conditions:
    - type: test
      status: "True"
      message: Everything is awesome!
      reason: Test`,
		},
		{
			name: "add condition to empty Kptfile",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status:
  conditions:
  - type: test
    status: "True"
    message: Everything is awesome!
    reason: Test`,
		},
		{
			name: "add condition to null status field",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status:`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status:
  conditions:
  - type: test
    status: "True"
    message: Everything is awesome!
    reason: Test`,
		},
		{
			name: "add condition to empty status field",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status: {}`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status:
  conditions:
  - type: test
    status: "True"
    message: Everything is awesome!
    reason: Test`,
		},
		{
			name: "add condition to bad status field",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status: bad`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status:
  conditions:
  - type: test
    status: "True"
    message: Everything is awesome!
    reason: Test`,
		},
		{
			name: "update existing half-empty condition",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status:
  conditions:
  - type: test`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status:
  conditions:
  - type: test
    status: "True"
    reason: Test
    message: Everything is awesome!`,
		},
		{
			name: "updating existing half-empty condition (one line)",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status: {conditions: [{type: test}]}`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status: {conditions: [{type: test, status: "True", reason: Test, message: Everything is awesome!}]}`,
		},
		{
			name: "updating existing condition (one line)",
			resources: map[string]string{
				"Kptfile": `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status: {conditions: [{type: test, status: "False", message: Everything is NOT awesome!, reason: TestFailed}]}`,
			},
			cond: kptfileapi.Condition{
				Type:    "test",
				Status:  "True",
				Reason:  "Test",
				Message: "Everything is awesome!",
			},
			expectedKptfile: `
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: example
status: {conditions: [{type: test, status: "True", message: Everything is awesome!, reason: Test}]}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			kptfile, err := NewKptfileFromPackage(tc.resources)
			assert.NilError(t, err, "failed to parse Kptfile")

			err = kptfile.SetTypedCondition(tc.cond)
			assert.NilError(t, err, "failed to set condition")

			err = kptfile.WriteToPackage(tc.resources)
			assert.NilError(t, err, "failed to write conditions back to Kptfile")
			assert.Equal(t, strings.TrimSpace(tc.expectedKptfile), strings.TrimSpace(tc.resources["Kptfile"]))

			gotCond, found := kptfile.GetTypedCondition("test")
			assert.Equal(t, true, found, "condition not found")
			assert.Equal(t, tc.cond, gotCond, "condition retrieved does not match the expected condition")

		})
	}
}
