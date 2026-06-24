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

package commands

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/heroku/color"
	"github.com/spf13/cobra"
)

const (
	DefaultFnPkg   = "https://github.com/kptdev/krm-functions-sdk.git/go/get-started@main"
	DefaultPkgName = "function"
)

func NewInitRunner(ctx context.Context) *InitRunner {
	r := &InitRunner{
		ctx: ctx,
	}
	r.Command = &cobra.Command{
		Use:     "init",
		Short:   "Initialize a new KRM function project",
		PreRunE: r.PreRunE,
		RunE:    r.RunE,
	}
	r.Command.Flags().StringVarP(&r.FnPkgPath, "fnPkg", "", DefaultFnPkg,
		"a kpt package that contains a basic KRM function source code to get start")
	return r
}

// InitRunner initializes a KRM function project from a scaffolded `kpt pkg`.
type InitRunner struct {
	ctx     context.Context
	Command *cobra.Command

	FnName    string
	FnPkgPath string
}

func (r *InitRunner) PreRunE(cmd *cobra.Command, args []string) error {
	if _, err := exec.LookPath("kpt"); err != nil {
		return fmt.Errorf("kfn requires that `kpt` is installed and on the PATH")
	}
	if len(args) == 0 {
		fmt.Printf(
			"Initializing the project in ./%v\nTip: You can customize the project name by running `kfn init <PROJECT>\n\n", DefaultPkgName)
		r.FnName = DefaultPkgName
	} else if len(args) == 1 {
		r.FnName = args[0]
	} else {
		return fmt.Errorf("no more than one argument is accepted, got %v", len(args))
	}
	return nil
}

func (r *InitRunner) RunE(cmd *cobra.Command, args []string) error {
	return r.GetFnPackage()
}

func (r *InitRunner) GetFnPackage() error {
	cmd := exec.Command("kpt", "pkg", "get", r.FnPkgPath, r.FnName) //nolint:gosec // CLI tool: args from user's own flags

	var out, errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	if err := cmd.Run(); err != nil {
		return err
	}
	color.Green("Now you can start writing your KRM function in ./%v/main.go\n", r.FnName)
	return nil
}
