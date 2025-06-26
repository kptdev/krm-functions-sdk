// Copyright 2023 The kpt Authors
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
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	// TODO: including this requires many dependency updates, at some point
	// we should do that so the CEL evaluation here is consistent with
	// K8s. There are a few other lines to uncomment in that case.
	//"k8s.io/apiserver/pkg/cel/library"
)

const (
	PkgContextVarName   = "package"
	UntypedItemsVarName = "items"
)

func (ko *KubeObject) ToUntyped() (interface{}, error) {
	return ko.obj.ToUntyped()
}

func (rl *ResourceList) ResolveCEL(celExpr string) (string, error) {
	inputs, err := rl.untypedCELInputs()
	if err != nil {
		return "", err
	}

	return evalExpr(celExpr, inputs)
}

func (rl *ResourceList) untypedCELInputs() (map[string]interface{}, error) {
	inputs := make(map[string]interface{})
	var items []interface{}

	for _, ko := range rl.Items {
		ut, err := ko.ToUntyped()
		if err != nil {
			return nil, err
		}
		items = append(items, ut)
	}

	inputs[UntypedItemsVarName] = items

	return inputs, nil
}

func evalExpr(expr string, inputs map[string]interface{}) (string, error) {
	prog, err := compileExpr(expr)
	if err != nil {
		return "", err
	}

	val, _, err := prog.Eval(inputs)
	if err != nil {
		return "", err
	}

	result, err := val.ConvertToNative(reflect.TypeOf(""))
	if err != nil {
		return "", err
	}

	s, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("expression returned non-string value: %v", result)
	}

	return s, nil
}

func unaryWithFunction(name string, adapter types.Adapter, path ...string) cel.EnvOption {
	fnname := fmt.Sprintf("with_%s", name)
	id := fmt.Sprintf("resourcelist_with_%s_string", name)
	return cel.Function(fnname,
		cel.MemberOverload(id, []*cel.Type{cel.ListType(cel.DynType), cel.StringType}, cel.DynType,
			cel.BinaryBinding(func(arg1, arg2 ref.Val) ref.Val {
				list := arg1.(traits.Lister)
				result := types.NewDynamicList(adapter, []any{})

				// loop through the list items to select the ones that match our criteria
				i := list.Iterator()
				for v := i.Next(); v != nil; v = i.Next() {
					mapper, ok := v.(traits.Mapper)
					if !ok {
						// if the entry is not a mapper, just skip it
						continue
					}
					// navigate through mappers for each field except the last
					for _, field := range path[:len(path)-1] {
						vv, ok := mapper.Find(adapter.NativeToValue(field))
						if !ok {
							// no value found for the field name
							// skip this list entry
							mapper = nil
							break
						}
						mapper, ok = vv.(traits.Mapper)
						if !ok {
							// value found for the field name is not a mapper
							// skip this list entry
							mapper = nil
							break
						}
					}
					if mapper == nil {
						// we could not successfully navigate the path, skip this list entry
						continue
					}
					// now pull the last field from the path; the result should be our
					// value you we want to check against
					testVal, ok := mapper.Find(adapter.NativeToValue(path[len(path)-1]))
					if !ok {
						// no such field, skip this list entry
						continue
					}
					// found the test value, compare it to the argument
					// and add it to the results list if found
					if testVal.Equal(arg2) == types.True {
						newResult := result.Add(types.NewDynamicList(adapter, []any{v}))
						result, ok = newResult.(traits.Lister)
						if !ok {
							continue
						}
					}
				}
				if result.Size() == types.IntOne {
					// if we found exactly one result, then return that result rather
					// than a list of one entry, avoiding the need to use the [0] indexing
					// notation
					return result.Get(types.IntZero)
				}

				// return the list result, so we can chain these
				return result
			}),
		),
	)
}

// compileExpr returns a compiled CEL expression.
func compileExpr(expr string) (cel.Program, error) {
	var opts []cel.EnvOption
	opts = append(opts, cel.HomogeneousAggregateLiterals())
	opts = append(opts, cel.EagerlyValidateDeclarations(true), cel.DefaultUTCTimeZone(true))

	// TODO: uncomment after updating to latest k8s
	//opts = append(opts, library.ExtensionLibs...)

	opts = append(opts, cel.Variable(UntypedItemsVarName, cel.ListType(cel.DynType)))

	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, err
	}

	env, err = env.Extend(
		unaryWithFunction("apiVersion", env.CELTypeAdapter(), "apiVersion"),
		unaryWithFunction("kind", env.CELTypeAdapter(), "kind"),
		unaryWithFunction("name", env.CELTypeAdapter(), "metadata", "name"),
		unaryWithFunction("namespace", env.CELTypeAdapter(), "metadata", "namespace"),
	)
	if err != nil {
		return nil, err
	}

	ast, issues := env.Compile(expr)
	if issues != nil {
		return nil, issues.Err()
	}

	_, err = cel.AstToCheckedExpr(ast)
	if err != nil {
		return nil, err
	}
	return env.Program(ast,
		cel.EvalOptions(cel.OptOptimize),
		// TODO: uncomment after updating to latest k8s
		//cel.OptimizeRegex(library.ExtensionLibRegexOptimizations...),
	)
}
