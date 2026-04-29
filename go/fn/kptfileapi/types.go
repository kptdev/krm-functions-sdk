// Copyright 2021,2026 The kpt Authors
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

// Package kptfileapi contains Kptfile API types.
//
// TEMPORARY COPY: These types are copied from github.com/kptdev/kpt/pkg/api/kptfile/v1
// to break a circular dependency (SDK depends on kpt, kpt depends on SDK).
//
// TARGET: Replace this package with an import from a central API repo
// (e.g. github.com/kptdev/api) once that repo is created.
package kptfileapi

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	KptFileName = "Kptfile"

	RevisionMetaDataFileName = ".KptRevisionMetadata"

	RevisionMetaDataKind = "KptRevisionMetadata"

	KptFileKind       = "Kptfile"
	KptFileGroup      = "kpt.dev"
	KptFileVersion    = "v1"
	KptFileAPIVersion = KptFileGroup + "/" + KptFileVersion
)

// KptFileGVK is the GroupVersionKind of Kptfile objects
func KptFileGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "kpt.dev",
		Version: "v1",
		Kind:    "Kptfile",
	}
}

// TypeMeta is the TypeMeta for KptFile instances.
var TypeMeta = yaml.ResourceMeta{
	TypeMeta: yaml.TypeMeta{
		APIVersion: KptFileAPIVersion,
		Kind:       KptFileKind,
	},
}

// KptFile contains information about a package managed with kpt.
type KptFile struct {
	yaml.ResourceMeta `yaml:",inline" json:",inline"`

	Upstream     *Upstream    `yaml:"upstream,omitempty" json:"upstream,omitempty"`
	UpstreamLock *Locator     `yaml:"upstreamLock,omitempty" json:"upstreamLock,omitempty"`
	Info         *PackageInfo `yaml:"info,omitempty" json:"info,omitempty"`
	Pipeline     *Pipeline    `yaml:"pipeline,omitempty" json:"pipeline,omitempty"`
	Inventory    *Inventory   `yaml:"inventory,omitempty" json:"inventory,omitempty"`
	Status       *Status      `yaml:"status,omitempty" json:"status,omitempty"`
}

type OriginType string

const (
	GitOrigin     OriginType = "git"
	GenericOrigin OriginType = "generic"
)

type UpdateStrategyType string

const (
	ResourceMerge      UpdateStrategyType = "resource-merge"
	FastForward        UpdateStrategyType = "fast-forward"
	ForceDeleteReplace UpdateStrategyType = "force-delete-replace"
	CopyMerge          UpdateStrategyType = "copy-merge"
)

func ToUpdateStrategy(strategy string) (UpdateStrategyType, error) {
	switch strategy {
	case string(ResourceMerge):
		return ResourceMerge, nil
	case string(FastForward):
		return FastForward, nil
	case string(ForceDeleteReplace):
		return ForceDeleteReplace, nil
	case string(CopyMerge):
		return CopyMerge, nil
	default:
		return "", fmt.Errorf("unknown update strategy %q", strategy)
	}
}

type Upstream struct {
	Type           OriginType         `yaml:"type,omitempty" json:"type,omitempty"`
	Git            *Git               `yaml:"git,omitempty" json:"git,omitempty"`
	UpdateStrategy UpdateStrategyType `yaml:"updateStrategy,omitempty" json:"updateStrategy,omitempty"`
}

type Git struct {
	Repo      string `yaml:"repo,omitempty" json:"repo,omitempty"`
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	Ref       string `yaml:"ref,omitempty" json:"ref,omitempty"`
}

type Locator struct {
	Type    OriginType   `yaml:"type,omitempty" json:"type,omitempty"`
	Git     *GitLock     `yaml:"git,omitempty" json:"git,omitempty"`
	Generic *GenericLock `yaml:"generic,omitempty" json:"generic,omitempty"`
}

type GitLock struct {
	Repo      string `yaml:"repo,omitempty" json:"repo,omitempty"`
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	Ref       string `yaml:"ref,omitempty" json:"ref,omitempty"`
	Commit    string `yaml:"commit,omitempty" json:"commit,omitempty"`
}

type GenericLock struct {
	StoreID         string `yaml:"storeID,omitempty" json:"storeID,omitempty"`
	ResourceID      string `yaml:"resourceID,omitempty" json:"resourceID,omitempty"`
	ResourceVersion string `yaml:"resourceVersion,omitempty" json:"resourceVersion,omitempty"`
}

type PackageInfo struct {
	Site           string          `yaml:"site,omitempty" json:"site,omitempty"`
	Emails         []string        `yaml:"emails,omitempty" json:"emails,omitempty"`
	License        string          `yaml:"license,omitempty" json:"license,omitempty"`
	LicenseFile    string          `yaml:"licenseFile,omitempty" json:"licenseFile,omitempty"`
	Description    string          `yaml:"description,omitempty" json:"description,omitempty"`
	Keywords       []string        `yaml:"keywords,omitempty" json:"keywords,omitempty"`
	Man            string          `yaml:"man,omitempty" json:"man,omitempty"`
	ReadinessGates []ReadinessGate `yaml:"readinessGates,omitempty" json:"readinessGates,omitempty"`
}

type ReadinessGate struct {
	ConditionType string `yaml:"conditionType" json:"conditionType"`
}

type Pipeline struct {
	Mutators   []Function `yaml:"mutators,omitempty" json:"mutators,omitempty"`
	Validators []Function `yaml:"validators,omitempty" json:"validators,omitempty"`
}

func (p *Pipeline) String() string {
	return fmt.Sprintf("%+v", *p)
}

func (p *Pipeline) IsEmpty() bool {
	if p == nil {
		return true
	}
	return len(p.Mutators) == 0 && len(p.Validators) == 0
}

type Function struct {
	Image      string            `yaml:"image,omitempty" json:"image,omitempty"`
	Exec       string            `yaml:"exec,omitempty" json:"exec,omitempty"`
	ConfigPath string            `yaml:"configPath,omitempty" json:"configPath,omitempty"`
	ConfigMap  map[string]string `yaml:"configMap,omitempty" json:"configMap,omitempty"`
	Name       string            `yaml:"name,omitempty" json:"name,omitempty"`
	Tag        string            `yaml:"tag,omitempty" json:"tag,omitempty"`
	Selectors  []Selector        `yaml:"selectors,omitempty" json:"selectors,omitempty"`
	Exclusions []Selector        `yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

type Selector struct {
	APIVersion  string            `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	Kind        string            `yaml:"kind,omitempty" json:"kind,omitempty"`
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type Inventory struct {
	Namespace   string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Name        string            `yaml:"name,omitempty" json:"name,omitempty"`
	InventoryID string            `yaml:"inventoryID,omitempty" json:"inventoryID,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

type Status struct {
	Conditions   []Condition   `yaml:"conditions,omitempty" json:"conditions,omitempty"`
	RenderStatus *RenderStatus `yaml:"renderStatus,omitempty" json:"renderStatus,omitempty"`
}

type RenderStatus struct {
	MutationSteps   []PipelineStepResult `yaml:"mutationSteps,omitempty" json:"mutationSteps,omitempty"`
	ValidationSteps []PipelineStepResult `yaml:"validationSteps,omitempty" json:"validationSteps,omitempty"`
	ErrorSummary    string               `yaml:"errorSummary,omitempty" json:"errorSummary,omitempty"`
}

type PipelineStepResult struct {
	Name           string       `yaml:"name,omitempty" json:"name,omitempty"`
	Image          string       `yaml:"image,omitempty" json:"image,omitempty"`
	ExecPath       string       `yaml:"exec,omitempty" json:"exec,omitempty"`
	ExecutionError string       `yaml:"executionError,omitempty" json:"executionError,omitempty"`
	Stderr         string       `yaml:"stderr,omitempty" json:"stderr,omitempty"`
	ExitCode       int          `yaml:"exitCode" json:"exitCode"`
	Results        []ResultItem `yaml:"results,omitempty" json:"results,omitempty"`
	ErrorResults   []ResultItem `yaml:"errorResults,omitempty" json:"errorResults,omitempty"`
}

type ResultItem struct {
	Message     string       `yaml:"message,omitempty" json:"message,omitempty"`
	Severity    string       `yaml:"severity,omitempty" json:"severity,omitempty"`
	ResourceRef *ResourceRef `yaml:"resourceRef,omitempty" json:"resourceRef,omitempty"`
	Field       *FieldRef    `yaml:"field,omitempty" json:"field,omitempty"`
	File        *FileRef     `yaml:"file,omitempty" json:"file,omitempty"`
}

type ResourceRef struct {
	APIVersion string `yaml:"apiVersion,omitempty" json:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty" json:"kind,omitempty"`
	Name       string `yaml:"name,omitempty" json:"name,omitempty"`
	Namespace  string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

type FieldRef struct {
	Path          string `yaml:"path,omitempty" json:"path,omitempty"`
	CurrentValue  string `yaml:"currentValue,omitempty" json:"currentValue,omitempty"`
	ProposedValue string `yaml:"proposedValue,omitempty" json:"proposedValue,omitempty"`
}

type FileRef struct {
	Path  string `yaml:"path,omitempty" json:"path,omitempty"`
	Index int    `yaml:"index,omitempty" json:"index,omitempty"`
}

type Condition struct {
	Type    string          `yaml:"type" json:"type"`
	Status  ConditionStatus `yaml:"status" json:"status"`
	Reason  string          `yaml:"reason,omitempty" json:"reason,omitempty"`
	Message string          `yaml:"message,omitempty" json:"message,omitempty"`
}

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

const (
	ConditionTypeRendered = "Rendered"
	ReasonRenderSuccess   = "RenderSuccess"
	ReasonRenderFailed    = "RenderFailed"
)

func NewRenderedCondition(status ConditionStatus, reason, message string) Condition {
	return Condition{
		Type:    ConditionTypeRendered,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
}

func ToCondition(value string) ConditionStatus {
	switch strings.ToLower(value) {
	case strings.ToLower(string(ConditionTrue)):
		return ConditionTrue
	case strings.ToLower(string(ConditionFalse)):
		return ConditionFalse
	default:
		return ConditionUnknown
	}
}
