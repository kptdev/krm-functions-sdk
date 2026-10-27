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
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This function generates Graphana configuration in the form of ConfigMap. It
// accepts Revision and ID as input.

func Example_generator() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(generate)); err != nil {
		os.Exit(1)
	}
}

// generate generates a ConfigMap.
func generate(rl *kubeobject.ResourceList) (bool, error) {
	if rl.FunctionConfig == nil {
		return false, kubeobject.ErrMissingFnConfig{}
	}

	revision, _, _ := rl.FunctionConfig.NestedString("data", "revision")
	id, _, _ := rl.FunctionConfig.NestedString("data", "id")
	js, err := fetchDashboard(revision, id)
	if err != nil {
		return false, fmt.Errorf("fetch dashboard: %v", err)
	}

	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-gen", rl.FunctionConfig.GetName()),
			Namespace: rl.FunctionConfig.GetNamespace(),
			Labels: map[string]string{
				"grafana_dashboard": "true",
			},
		},
		Data: map[string]string{
			fmt.Sprintf("%v.json", rl.FunctionConfig.GetName()): fmt.Sprintf("%q", js),
		},
	}
	return true, rl.UpsertObjectToItems(cm, nil, false)
}

func fetchDashboard(revision, id string) (string, error) {
	url := fmt.Sprintf("https://grafana.com/api/dashboards/%s/revisions/%s/download", id, revision)
	resp, err := http.Get(url) // nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
