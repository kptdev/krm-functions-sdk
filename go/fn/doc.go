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

/*
Package fn provides the SDK to write KRM functions.

# Before you start

This fn SDK requires some basic KRM function Specification knowledge. To make the best usage of your time, we recommend
you to be familiar with "ResourceList" before moving forward.

	The KRM Function Specification, or "ResourceList", defines the standards of the inter-process communication between
	the orchestrator (i.e. kpt CLI) and functions.

See KRM Function Specification reference in https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md

# KRM Function

A KRM function can mutate and/or validate Kubernetes resources in a ResourceList.

KubeObject, SubObject, and ResourceList live in github.com/kptdev/kpt/api/kubeobject.
The ResourceList type maps to the ResourceList in the function spec.

Read more about how to use KRM functions in https://kpt.dev/book/04-using-functions/

Read more about how to develop a KRM function in https://kpt.dev/book/05-developing-functions/

A general workflow is:
 1. Reads the "ResourceList" object from STDIN.
 2. Gets the function configs from the "ResourceList.FunctionConfig".
 3. Mutate or validate the Kubernetes YAML resources from the "ResourceList.Items" field with the function configs.
 4. Writes the modified "ResourceList" to STDOUT.
 5. Write function message to "ResourceList.Results" with severity "Info", "Warning" or "Error"

# KubeObject

KubeObject is the basic unit for operating on KRM resources. It, along with SubObject
and the helpers for reading and writing resources, is provided by
github.com/kptdev/kpt/api/kubeobject.

In "AsMain", both "Items" and "FunctionConfig" are converted to KubeObject values.

# AsMain

"AsMain" is the main entrypoint. In most cases, you only need to provide the mutator or validation logic and have AsMain
handles the ResourceList parsing, KRM resource field type detection, read from STDIN and write to STDOUT.

"AsMain" accepts a kubeobject.ResourceListProcessor or a Runner.

See github.com/kptdev/krm-functions-sdk/go/fn/examples for detailed usage.
*/
package fn
