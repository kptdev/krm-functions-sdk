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
	"github.com/kptdev/kpt/api/kubeobject"
	corev1 "k8s.io/api/core/v1"
)

var (
	deployment kubeobject.KubeObject
	configMap  kubeobject.KubeObject
)

func Example_kubeObjectMutatePrimitiveField() {
	spec := deployment.GetMap("spec")
	replicas := spec.GetInt("replicas")
	// mutate the replicas variable

	err := spec.SetNestedInt(int(replicas))
	if err != nil {
		panic(err)
	}
}

func Example_kubeObjectMutatePrimitiveSlice() {
	finalizers, _, _ := deployment.NestedStringSlice("metadata", "finalizers")
	// mutate the finalizers slice

	err := deployment.SetNestedStringSlice(finalizers, "metadata", "finalizers")
	if err != nil {
		panic(err)
	}
}

func Example_kubeObjectMutatePrimitiveMap() {
	data, _, _ := configMap.NestedStringMap("data")
	// mutate the data map

	err := deployment.SetNestedStringMap(data, "data")
	if err != nil { /* do something */
	}
}

func Example_kubeObjectMutateStrongTypedField() {
	var newPodTemplate corev1.PodTemplate
	curPodTemplate := configMap.GetMap("spec").GetMap("template")
	// Assign the current PodTemplate value to newPodTemplate
	// Use As to AsMain handles the errors.
	err := curPodTemplate.As(&newPodTemplate)
	if err != nil {
		panic(err)
	}
	// mutate the newPodTemplate object
	err = deployment.SetNestedField(newPodTemplate, "spec", "template")
	if err != nil { /* do something */
		panic(err)
	}
}

func Example_kubeObjectMutateStrongTypedSlice() {
	var containers []corev1.Container
	found, err := deployment.NestedResource(&containers, "spec", "template", "spec", "containers")
	if err != nil { /* do something */
	}
	if !found { /* do something */
	}

	// mutate the podTemplate object

	err = deployment.SetNestedField(containers, "spec", "template", "spec", "containers")
	if err != nil { /* do something */
	}
}
