// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/fn/framework/frameworktestutil"
)

func TestRun(t *testing.T) {
	prc := frameworktestutil.CommandResultsChecker{
		Command: buildCmd,
	}
	prc.Assert(t)
}
