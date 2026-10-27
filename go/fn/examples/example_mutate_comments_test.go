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
	"strings"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// In this example, we mutate line comments for field metadata.name.
// Some function may want to store some information in the comments (e.g.
// apply-setters function: https://catalog.kpt.dev/apply-setters/v0.2/)

func Example_dMutateComments() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(mutateComments)); err != nil {
		os.Exit(1)
	}
}

func mutateComments(rl *kubeobject.ResourceList) (bool, error) {
	for i := range rl.Items {
		lineComment, found, err := rl.Items[i].LineComment("metadata", "name")
		if err != nil {
			return false, err
		}
		if !found {
			return true, nil
		}

		if strings.TrimSpace(lineComment) == "" {
			lineComment = "bar-system"
		} else {
			lineComment = strings.Replace(lineComment, "foo", "bar", -1)
		}
		if err = rl.Items[i].SetLineComment(lineComment, "metadata", "name"); err != nil {
			return false, err
		}
	}
	return true, nil
}
