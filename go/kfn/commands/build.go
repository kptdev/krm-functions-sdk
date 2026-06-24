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
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/heroku/color"
	"github.com/spf13/cobra"
)

//go:embed embed/Dockerfile
var f embed.FS
var execCmdFn = execCmd
var execLookPathFn = exec.LookPath

const (
	// Builder Type
	Ko     = "ko"
	Docker = "docker"

	// Docker constant variables
	Image                 = "function:latest"
	DockerfilePath        = "Dockerfile"
	BuiltinDockerfilePath = "embed/Dockerfile"

	// Ko constant variables
	KoDockerRepoEnvVar = "KO_DOCKER_REPO"
	KoLocalRepo        = "ko.local"

	build = "build"
)

func NewBuildRunner(ctx context.Context) *BuildRunner {
	r := &BuildRunner{
		ctx:    ctx,
		Ko:     &KoBuilder{},
		Docker: &DockerBuilder{},
	}
	r.Command = &cobra.Command{
		Use:   build,
		Short: "build your KRM function to a container image",
		RunE:  r.RunE,
	}
	r.Command.Flags().StringVarP(&r.BuilderType, "builder", "b", Ko,
		"the image builder. `ko` is the default builder, which requires `go build`; `docker` is accepted, and "+
			" requires you to have docker installed and running")
	r.Command.Flags().StringVarP(&r.Docker.Image, "image", "i", Image,
		fmt.Sprintf("the image (with tag), default to %v", Image))
	r.Command.Flags().StringVarP(&r.Docker.DockerfilePath, "dockerfile", "f", "",
		"path to the Dockerfile. If not given, using a default builtin Dockerfile")
	r.Command.Flags().StringVarP(&r.Ko.Repo, "repo", "r", "",
		"the image repo. default to ko.local")
	r.Command.Flags().StringVarP(&r.Ko.Tag, "tag", "t", "latest",
		"the ko image tag")
	// TODO: Docker CLI uses `--tag` flag to refer to "image:tag", which could be confusing but broadly accepted.
	// We should better guide users on how to use "tag" and "image" flags for kfn.
	// Here we use "tag" for ko <tag> (same as `ko build --tag`) and "image" for docker <image:tag> (same as `docker build --tag`)
	return r
}

type BuildRunner struct {
	ctx     context.Context
	Command *cobra.Command

	BuilderType string
	Tag         string
	Ko          *KoBuilder
	Docker      *DockerBuilder
}

type Builder interface {
	Build() error
	Validate() error
}

type DockerBuilder struct {
	Image          string
	DockerfilePath string
}

type KoBuilder struct {
	Repo string
	Tag  string
}

func (r *BuildRunner) RunE(cmd *cobra.Command, args []string) error {
	var builder Builder
	switch r.BuilderType {
	case Docker:
		builder = r.Docker
	case Ko:
		builder = r.Ko
	}
	if err := builder.Validate(); err != nil {
		return err
	}
	return builder.Build()
}

func (r *DockerBuilder) Build() error {
	args := []string{build, ".", "-f", r.DockerfilePath, "--tag", r.Image}
	err := execCmdFn(nil, "docker", args...)
	if err != nil {
		return err
	}
	color.Green("Image %v built successfully. Now you can publish the image", r.Image)
	return nil
}

func (r *DockerBuilder) Validate() error {
	if err := r.validateDockerInstalled(); err != nil {
		return err
	}
	if r.dockerfileExists() {
		return nil
	}
	return r.createDockerfile()
}

func (r *DockerBuilder) validateDockerInstalled() error {
	_, err := execLookPathFn("docker")
	if err != nil {
		return fmt.Errorf("kfn requires that `docker` is installed and on the PATH")
	}
	return nil
}

func (r *DockerBuilder) dockerfileExists() bool {
	if r.DockerfilePath == "" {
		r.DockerfilePath = DockerfilePath
	}
	if _, err := os.Stat(r.DockerfilePath); errors.Is(err, os.ErrNotExist) {
		color.Yellow("not find %v, using the builtin default Dockerfile instead...", r.DockerfilePath)
		return false
	}
	return true
}

func (r *DockerBuilder) createDockerfile() error {
	dockerfileContent, err := f.ReadFile(BuiltinDockerfilePath)
	if err != nil {
		return err
	}
	if err = os.WriteFile(DockerfilePath, dockerfileContent, 0600); err != nil {
		return err
	}
	fmt.Println("created Dockerfile")
	return nil
}

func (r *KoBuilder) GuaranteeKoInstalled() error {
	_, err := execLookPathFn("ko")
	if err == nil {
		return nil
	}
	gobin := os.Getenv("GOBIN")
	if gobin == "" && os.Getenv("GOPATH") != "" {
		gobin = os.Getenv("GOPATH") + "/bin"
	}
	if gobin == "" && os.Getenv("HOME") != "" {
		gobin = os.Getenv("HOME") + "/go/bin"
	}
	var envs []string
	if gobin != "" {
		envs = []string{"GOBIN" + "=" + gobin}
	}
	if err = execCmdFn(envs, "go", "install", "github.com/google/ko@latest"); err != nil {
		return err
	}
	fmt.Println("successfully installed ko")
	return nil
}
func (r *KoBuilder) Build() error {
	args := []string{build, "-B", "--tags", r.Tag}
	envs := []string{KoDockerRepoEnvVar + "=" + r.Repo}
	err := execCmdFn(envs, "ko", args...)
	if err != nil {
		return err
	}

	if r.Repo == KoLocalRepo {
		color.Green("Image built successfully. Now you can publish the image")
	} else {
		color.Green("Image built and pushed successfully")
	}
	return nil
}

func (r *KoBuilder) Validate() error {
	if err := r.GuaranteeKoInstalled(); err != nil {
		return err
	}
	// Find KO_DOCKER_REPO value from multiple places for `ko build`.
	if r.Repo != "" {
		return nil
	}
	if repo, ok := os.LookupEnv(KoDockerRepoEnvVar); ok {
		r.Repo = repo
		return nil
	}
	r.Repo = "ko.local"
	return nil
}

func execCmd(envs []string, name string, args ...string) error {
	cmd := exec.Command(name, args...) //nolint:gosec // CLI tool: args constructed internally from user's own flags
	if len(envs) != 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, envs...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
