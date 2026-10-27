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
	"bytes"
	"fmt"
	"strings"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// This example removes all resources as a builtin function by using fn.Execute.

type RemoveAllResources struct{}

func (*RemoveAllResources) Process(rl *kubeobject.ResourceList) (bool, error) {
	rl.Items = nil
	return true, nil
}

func Example_builtinFunction() {
	reader := strings.NewReader(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- kind: Deployment
  metadata:
    name: my-deploy
- kind: Service
  metadata:
    name: my-service
functionConfig:
  apiVersion: fn.kpt.dev/v1alpha1
  kind: RemoveAllResources
  metadata:
    name: fn-config`)

	var writer bytes.Buffer
	err := fn.Execute(&RemoveAllResources{}, reader, &writer)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(writer.String())

	// Output:
	// apiVersion: config.kubernetes.io/v1
	// kind: ResourceList
	// items: []
	// functionConfig:
	//   apiVersion: fn.kpt.dev/v1alpha1
	//   kind: RemoveAllResources
	//   metadata:
	//     name: fn-config
	//
}
