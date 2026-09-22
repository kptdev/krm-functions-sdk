// Copyright 2022, 2026 The kpt Authors
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

package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

// greetingAnnotation is the annotation this example stamps onto every resource.
const greetingAnnotation = "example.kpt.dev/greeting"

var _ fn.Runner = &HelloWorld{}

// HelloWorld is the functionConfig for this "hello world" example. The struct
// name (HelloWorld) is used as the functionConfig `kind`, and each exported
// field is populated from the matching functionConfig key via its JSON tag.
//
// For example, this functionConfig sets both fields below:
//
//	apiVersion: fn.kpt.dev/v1alpha1
//	kind: HelloWorld
//	greeting: Hello
//	name: world
//
// TODO: Rename this struct to your functionConfig "kind" and replace the fields
// with the configuration your function needs.
type HelloWorld struct {
	// Greeting is the salutation to use. Defaults to "Hello" when empty.
	Greeting string `json:"greeting,omitempty"`
	// Name is who to greet. Defaults to "world" when empty.
	Name string `json:"name,omitempty"`
}

// Run is the main function logic.
//   - `items` are the resources parsed from "ResourceList.items". You may modify
//     existing items but not add or remove them (use a ResourceListProcessor for that).
//   - `functionConfig` is "ResourceList.functionConfig"; its values have already
//     been unmarshaled into the receiver's fields (r.Greeting and r.Name here).
//   - `results` is "ResourceList.results", where you report info/warning/error messages.
//
// This example builds a greeting from the functionConfig and stamps it as an
// annotation on every resource.
func (r *HelloWorld) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	greeting := r.Greeting
	if greeting == "" {
		greeting = "Hello"
	}
	name := r.Name
	if name == "" {
		name = "world"
	}
	message := fmt.Sprintf("%s, %s!", greeting, name)

	for _, obj := range items {
		if err := obj.SetAnnotation(greetingAnnotation, message); err != nil {
			results.ErrorE(err)
		}
	}
	results.Infof("greeted %d resource(s) with %q", len(items), message)
	return results.ExitCode() == 0
}

func main() {
	runner := fn.WithContext(context.Background(), &HelloWorld{})
	if err := fn.AsMain(runner, fn.WithDocs(readme, metadata)); err != nil {
		os.Exit(1)
	}
}
