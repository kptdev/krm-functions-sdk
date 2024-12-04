// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/kio"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/kio/filters"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"
)

func TestGrepFilter_Filter(t *testing.T) {
	in := `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx
  annotations:
    app: nginx
  name: bar
spec:
  replicas: 3
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`
	out := &bytes.Buffer{}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{filters.GrepFilter{Path: []string{"metadata", "name"}, Value: "foo"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Service
metadata:
  name: foo
  annotations:
    app: nginx
spec:
  selector:
    app: nginx
`, out.String()) {
		t.FailNow()
	}

	out = &bytes.Buffer{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{filters.GrepFilter{Path: []string{"kind"}, Value: "Deployment"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx2
  name: foo
  annotations:
    app: nginx2
spec:
  replicas: 1
---
kind: Deployment
metadata:
  labels:
    app: nginx
  annotations:
    app: nginx
  name: bar
spec:
  replicas: 3
`, out.String()) {
		t.FailNow()
	}

	out = &bytes.Buffer{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{filters.GrepFilter{Path: []string{"spec", "replicas"}, Value: "3"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, `kind: Deployment
metadata:
  labels:
    app: nginx
  annotations:
    app: nginx
  name: bar
spec:
  replicas: 3
`, out.String()) {
		t.FailNow()
	}

	out = &bytes.Buffer{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(in)}},
		Filters: []kio.Filter{filters.GrepFilter{Path: []string{"spec", "not-present"}, Value: "3"}},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: out}},
	}.Execute()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, ``, out.String()) {
		t.FailNow()
	}
}

func TestGrepFilter_init(t *testing.T) {
	assert.Equal(t, filters.GrepFilter{}, filters.Filters["GrepFilter"]())
}

func TestGrepFilter_error(t *testing.T) {
	v, err := filters.GrepFilter{Path: []string{"metadata", "name"},
		Value: "foo"}.Filter([]*yaml.RNode{{}})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Nil(t, v)
}
