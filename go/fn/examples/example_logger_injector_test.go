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
	corev1 "k8s.io/api/core/v1"
	yaml2 "sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// In this example, we implement a function that injects a logger as a sidecar
// container in workload APIs.

func Example_loggeInjector() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(injectLogger)); err != nil {
		os.Exit(1)
	}
}

// injectLogger injects a logger container into the workload API resources.
// generate implements the gokrmfn.KRMFunction interface.
func injectLogger(rl *kubeobject.ResourceList) (bool, error) {
	var li LoggerInjection
	if err := rl.FunctionConfig.As(&li); err != nil {
		return false, err
	}
	for i, obj := range rl.Items.Where(hasDesiredGVK) {
		var containers []corev1.Container
		found, err := obj.NestedResource(&containers, "spec", "template", "spec", "containers")
		if err != nil || !found {
			return found, err
		}
		foundTargetContainer := false
		for j, container := range containers {
			if container.Name == li.ContainerName {
				containers[j].Image = li.ImageName
				foundTargetContainer = true
				break
			}
		}
		if !foundTargetContainer {
			c := corev1.Container{
				Name:  li.ContainerName,
				Image: li.ImageName,
			}
			containers = append(containers, c)
		}
		if err = rl.Items[i].SetNestedField(containers, "spec", "template", "spec", "containers"); err != nil {
			return false, nil
		}
	}
	return true, nil
}

// LoggerInjection is type definition of the functionConfig.
type LoggerInjection struct {
	yaml2.ResourceMeta `json:",inline" yaml:",inline"`

	ContainerName string `json:"containerName" yaml:"containerName"`
	ImageName     string `json:"imageName" yaml:"imageName"`
}
