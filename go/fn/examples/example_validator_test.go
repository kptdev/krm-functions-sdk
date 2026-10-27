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

package example

import (
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// This example implements a function that validate resources to ensure
// spec.template.spec.securityContext.runAsNonRoot is set in workload APIs.
func Example_validator() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(validator)); err != nil {
		os.Exit(1)
	}
}

func validator(rl *kubeobject.ResourceList) (bool, error) {
	var results kubeobject.Results
	for _, obj := range rl.Items.Where(hasDesiredGVK) {
		var runAsNonRoot bool
		if _, err := obj.NestedResource(&runAsNonRoot, "spec", "template", "spec", "securityContext", "runAsNonRoot"); err != nil {
			return false, err
		}
		if !runAsNonRoot {
			results = append(results, kubeobject.ConfigObjectResult("`spec.template.spec.securityContext.runAsNonRoot` must be set to true", obj, kubeobject.Error))
		}
	}
	return true, results
}
