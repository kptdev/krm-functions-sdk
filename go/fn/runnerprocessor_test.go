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

package fn

import (
	"context"
	"fmt"
	"testing"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/stretchr/testify/assert"
)

var _ Runner = &SetTest{}

type SetTestWithArgs interface {
	GetArgs() string
}

type SetTest struct {
	Arg map[string]string
}

func (*SetTest) Run(*Context, *kubeobject.KubeObject, kubeobject.KubeObjects, *kubeobject.Results) bool {
	return true
}

func (s *SetTest) GetArgs() string {
	return fmt.Sprintf("%v", s.Arg)
}

type SetTestNoMapString struct {
	Arg string `yaml:"args" json:"args"`
}

func (*SetTestNoMapString) Run(*Context, *kubeobject.KubeObject, kubeobject.KubeObjects, *kubeobject.Results) bool {
	return true
}

func (s *SetTestNoMapString) GetArgs() string {
	return s.Arg
}

func TestProcess(t *testing.T) {
	testdata := map[string]struct {
		resourceList []byte
		expectedOk   bool
		expectedErr  string
	}{
		"functionConfig is empty": {
			resourceList: []byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig: {}
`),
			expectedOk:  true,
			expectedErr: "[info]: `FunctionConfig` is not given",
		},
		"functionConfig is create by kpt but actually is empty": {
			resourceList: []byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig: 
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: function-input
  data: {}
`),
			expectedOk:  true,
			expectedErr: "[info]: `FunctionConfig` is not given",
		},
		"functionConfig error": {
			resourceList: []byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig: 
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: function-input
  data: wrong-type
`),
			expectedOk:  false,
			expectedErr: "[error]: Resource(apiVersion=, kind=ConfigMap) has unmatched field type \"map[string]string\" in fieldpath .data",
		},
		"functionConfig pass": {
			resourceList: []byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
functionConfig: 
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: function-input
  data: 
    k1: v1
    k2: v2
`),
			expectedOk:  true,
			expectedErr: "",
		},
	}
	for description, test := range testdata {
		fnConfig := &SetTest{}
		r := runnerProcessor{ctx: context.TODO(), fnRunner: fnConfig}
		rl, _ := kubeobject.ParseResourceList(test.resourceList)
		actualOk, _ := r.Process(rl)
		assert.Equal(t, actualOk, test.expectedOk, description)
		assert.Equal(t, test.expectedErr, rl.Results.String(), description)
	}
}

func TestRunnerConfig(t *testing.T) {
	testdata := map[string]struct {
		fnConfig             []byte
		expectedErr          string
		expectedArgsToString string
		runner               SetTestWithArgs
	}{
		"functionConfig is ConfigMap, invalid data": {
			fnConfig: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data: wrong-type
`),
			runner:      &SetTest{},
			expectedErr: "Resource(apiVersion=, kind=ConfigMap) has unmatched field type \"map[string]string\" in fieldpath .data",
		},
		"functionConfig is ConfigMap, value assigned to Runner arg": {
			fnConfig: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data: 
    k1: v1
    k2: v2
`),
			runner:      &SetTest{},
			expectedErr: "",
			expectedArgsToString: fmt.Sprintf("%v", map[string]string{
				"k1": "v1",
				"k2": "v2",
			}),
		},
		"functionConfig is ConfigMap, Runner does not have available arg to assign data": {
			fnConfig: []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data: 
    k1: v1
    k2: v2
`),
			runner: &SetTestNoMapString{},
			expectedErr: "unable to assign the given ConfigMap `.data` to FunctionConfig SetTestNoMapString. " +
				"please make sure the SetTestNoMapString has a field of type map[string]string",
			expectedArgsToString: "",
		},
		"functionConfig is custom kind, validate the GVK ": {
			fnConfig: []byte(`
apiVersion: badgroup/v1alpha1
kind: SetTestNoMapString
metadata:
  name: test
`),
			runner:               &SetTestNoMapString{},
			expectedErr:          "unknown FunctionConfig `SetTestNoMapString.badgroup`, expect `SetTestNoMapString.fn.kpt.dev` or `ConfigMap.v1`",
			expectedArgsToString: "",
		},
		"functionConfig is custom kind, assign the value to runner": {
			fnConfig: []byte(`
apiVersion: fn.kpt.dev/v1alpha1
kind: SetTestNoMapString
metadata:
  name: test
args: test
`),
			runner:               &SetTestNoMapString{},
			expectedErr:          "",
			expectedArgsToString: "test",
		},
	}
	for description, test := range testdata {
		r := runnerProcessor{ctx: context.TODO(), fnRunner: test.runner.(Runner)}
		fn, _ := kubeobject.ParseKubeObject(test.fnConfig)
		actualErr := r.config(fn)
		if test.expectedErr != "" {
			assert.EqualError(t, actualErr, test.expectedErr, description)
		} else {
			assert.Equal(t, test.expectedArgsToString, test.runner.GetArgs(), description)
		}
	}
}
