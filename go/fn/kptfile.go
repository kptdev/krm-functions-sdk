// Copyright 2024 The kpt and Nephio Authors
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
	"sort"

	kptfileapi "github.com/kptdev/kpt/pkg/api/kptfile/v1"
)

const (
	statusFieldName     = "status"
	conditionsFieldName = "conditions"
)

var (
	BoolToConditionStatus = map[bool]kptfileapi.ConditionStatus{
		true:  kptfileapi.ConditionTrue,
		false: kptfileapi.ConditionFalse,
	}
)

// Kptfile provides an API to manipulate the Kptfile of a kpt package
type Kptfile struct {
	Obj *KubeObject
}

// NewKptfileFromKubeObjectList creates a KptfileObject by finding it in the given KubeObjects list
func NewKptfileFromKubeObjectList(objs KubeObjects) (*Kptfile, error) {
	var ret Kptfile
	ret.Obj = objs.GetRootKptfile()
	if ret.Obj == nil {
		return nil, fmt.Errorf("the Kptfile object is missing from the package")
	}
	return &ret, nil
}

// NewKptfileFromPackage creates a KptfileObject from the resource (YAML) files of a package
func NewKptfileFromPackage(resources map[string]string) (*Kptfile, error) {
	kptfileStr, found := resources[kptfileapi.KptFileName]
	if !found {
		return nil, fmt.Errorf("%s is missing from the package", kptfileapi.KptFileName)
	}

	kos, err := ReadKubeObjectsFromFile(kptfileapi.KptFileName, kptfileStr)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse %s from package: %w", kptfileapi.KptFileName, err)
	}
	return NewKptfileFromKubeObjectList(kos)
}

func (kf *Kptfile) WriteToPackage(resources map[string]string) error {
	if kf == nil || kf.Obj == nil {
		return fmt.Errorf("attempt to write empty Kptfile to the package")
	}
	kptfileStr, err := WriteKubeObjectsToString(KubeObjects{kf.Obj})
	if err != nil {
		return err
	}
	resources[kptfileapi.KptFileName] = kptfileStr
	return nil
}

func (kf *Kptfile) String() string {
	if kf.Obj == nil {
		return ""
	}
	kptfileStr, _ := WriteKubeObjectsToString(KubeObjects{kf.Obj})
	return kptfileStr
}

// Status returns with the Status field of the Kptfile as a SubObject
// If the Status field doesn't exist, it is added.
func (kf *Kptfile) Status() *SubObject {
	return kf.Obj.UpsertMap(statusFieldName)
}

func (kf *Kptfile) Conditions() SliceSubObjects {
	return kf.Status().GetSlice(conditionsFieldName)
}

func (kf *Kptfile) SetConditions(conditions SliceSubObjects) error {
	sort.SliceStable(conditions, func(i, j int) bool {
		return conditions[i].GetString("type") < conditions[j].GetString("type")
	})
	return kf.Status().SetSlice(conditions, conditionsFieldName)
}

// TypedConditions returns with (a copy of) the list of current conditions of the kpt package
func (kf *Kptfile) TypedConditions() []kptfileapi.Condition {
	statusObj := kf.Obj.GetMap(statusFieldName)
	if statusObj == nil {
		return nil
	}
	var status kptfileapi.Status
	err := statusObj.As(&status)
	if err != nil {
		return nil
	}
	return status.Conditions
}

// GetTypedCondition returns with the condition whose type is `conditionType` as its first return value, and
// whether the component exists or not as its second return value
func (kf *Kptfile) GetTypedCondition(conditionType string) (kptfileapi.Condition, bool) {
	for _, cond := range kf.TypedConditions() {
		if cond.Type == conditionType {
			return cond, true
		}
	}
	return kptfileapi.Condition{}, false
}

// SetTypedCondition creates or updates the given condition using the Type field as the primary key
func (kf *Kptfile) SetTypedCondition(condition kptfileapi.Condition) error {
	conditions := kf.Conditions()
	for _, conditionSubObj := range conditions {
		if conditionSubObj.GetString("type") == condition.Type {
			// use the SetNestedString methods as opposed to SetNestedStringMap
			// in order to keep the order of new fields deterministic
			if err := conditionSubObj.SetNestedString(string(condition.Status), "status"); err != nil {
				return err
			}
			if err := conditionSubObj.SetNestedString(condition.Reason, "reason"); err != nil {
				return err
			}
			if err := conditionSubObj.SetNestedString(condition.Message, "message"); err != nil {
				return err
			}
			return kf.SetConditions(conditions)
		}
	}
	ko, err := NewFromTypedObject(condition)
	if err != nil {
		return err
	}
	conditions = append(conditions, &ko.SubObject)
	return kf.SetConditions(conditions)
}

// DeleteConditionByType deletes all conditions with the given type
func (kf *Kptfile) DeleteConditionByType(conditionType string) error {
	oldConditions, found, err := kf.Obj.NestedSlice(conditionsFieldName)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	newConditions := make([]*SubObject, 0, len(oldConditions))
	for _, c := range oldConditions {
		if c.GetString("type") != conditionType {
			newConditions = append(newConditions, c)
		}
	}
	return kf.SetConditions(newConditions)
}

func (kf *Kptfile) AddReadinessGates(gates []kptfileapi.ReadinessGate) error {
	info := kf.Obj.UpsertMap("info")
	gateObjs := info.GetSlice("readinessGates")
	for _, gate := range gates {
		// check if readiness gate already exists
		found := false
		for _, gateObj := range gateObjs {
			if gateObj.GetString("conditionType") == gate.ConditionType {
				found = true
				break
			}
		}
		// add if not found
		if !found {
			ko, err := NewFromTypedObject(gate)
			if err != nil {
				return err
			}
			gateObjs = append(gateObjs, &ko.SubObject)
		}
	}
	if err := info.SetSlice(gateObjs, "readinessGates"); err != nil {
		return err
	}
	return nil
}

func (kf *Kptfile) AddMutationFunction(fn *kptfileapi.Function) error {
	pipeline := kf.Obj.UpsertMap("pipeline")
	mutators := pipeline.GetSlice("mutators")
	ko, err := NewFromTypedObject(fn)
	if err != nil {
		return fmt.Errorf("failed to add mutator function (%s) to Kptfile: %w", fn.Image, err)
	}
	mutators = append(mutators, &ko.SubObject)
	if err := pipeline.SetSlice(mutators, "mutators"); err != nil {
		return err
	}
	return nil
}
