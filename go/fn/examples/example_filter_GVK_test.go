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

package example_test

import (
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// This example implements a function that updates the replicas field for all deployments.

func Example_filterGVK() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(updateReplicas)); err != nil {
		os.Exit(1)
	}
}

// updateReplicas sets a field in resources selecting by GVK.
func updateReplicas(rl *kubeobject.ResourceList) (bool, error) {
	if rl.FunctionConfig == nil {
		return false, kubeobject.ErrMissingFnConfig{}
	}
	var replicas int
	found, err := rl.FunctionConfig.NestedResource(&replicas, "replicas")
	if err != nil || !found {
		return found, err
	}
	for i := range rl.Items.Where(kubeobject.IsGVK("apps", "v1", "Deployment")) {
		if err := rl.Items[i].SetNestedField(replicas, "spec", "replicas"); err != nil {
			return false, err
		}
	}
	return true, nil
}
