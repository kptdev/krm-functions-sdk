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

import "github.com/kptdev/kpt/api/kubeobject"

const (
	apps = "apps"
	v1   = "v1"
)

func hasDesiredGVK(obj *kubeobject.KubeObject) bool {
	return obj.IsGVK(apps, v1, "Deployment") || obj.IsGVK(apps, v1, "StatefulSet") ||
		obj.IsGVK(apps, v1, "DaemonSet") || obj.IsGVK(apps, v1, "ReplicaSet")
}
