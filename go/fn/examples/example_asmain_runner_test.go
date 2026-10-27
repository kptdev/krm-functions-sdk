// Copyright 2022 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package example

import (
	"context"
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

var _ fn.Runner = &SetLabels{}

type SetLabels struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// Run is the main function logic.
// `ctx` provides easy methods to add info, error or warning result to `ResourceList.Results`.
// `items` is parsed from the STDIN "ResourceList.Items".
// `functionConfig` is from the STDIN "ResourceList.FunctionConfig". The value has been assigned to the r.Labels
//
//	the functionConfig is validated to have kind "SetLabels" and apiVersion "fn.kpt.dev/v1alpha1"
func (r *SetLabels) Run(ctx *fn.Context, functionConfig *kubeobject.KubeObject, items kubeobject.KubeObjects, results *kubeobject.Results) bool {
	for _, o := range items {
		for k, newLabel := range r.Labels {
			err := o.SetLabel(k, newLabel)
			if err != nil {
				results.ErrorE(err)
				// continue even if error occurs, we want the final results to show all errors.
			}
		}
	}
	if results.ExitCode() != 1 {
		results.Infof("updated labels")
	}
	return true
}

// This example uses a SetLabels object, which implements `Runner.Run` methods.
//
// The input from ./data/setlabels-resourcelist.yaml:
// apiVersion: config.kubernetes.io/v1
// kind: ResourceList
// items:
//   - apiVersion: v1
//     kind: Service
//     metadata:
//     name: example
//
// functionConfig:
//
//	apiVersion: fn.kpt.dev/v1alpha1
//	kind: SetLabels
//	metadata:
//	  name: setlabel-fn-config
func Example_asMain() {
	file, _ := os.Open("./data/setlabels-resourcelist.yaml")
	defer file.Close()
	os.Stdin = file
	ctx := context.TODO()
	if err := fn.AsMain(fn.WithContext(ctx, &SetLabels{})); err != nil {
		os.Exit(1)
	}
	// Output:
	// apiVersion: config.kubernetes.io/v1
	// kind: ResourceList
	// items:
	// - apiVersion: v1
	//   kind: Service
	//   metadata:
	//     name: example
	// functionConfig:
	//   apiVersion: fn.kpt.dev/v1alpha1
	//   kind: SetLabels
	//   metadata:
	//     name: setlabel-fn-config
	// results:
	// - message: updated labels
	//   severity: info
}
