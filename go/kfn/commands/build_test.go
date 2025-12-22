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

package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	testcases := map[string]struct {
		args []string
		// expected args is key, and expected return is value
		cmdExpected []string
		// expected env var is key, and existence is value
		lookPathExpected   map[string]bool
		expectedError      string
		dockerfileExpected bool
	}{
		"default build is ko, ko already exists": {
			args: []string{""},
			cmdExpected: []string{
				"KO_DOCKER_REPO=ko.local ko build -B --tags latest",
			},
			lookPathExpected: map[string]bool{
				"ko": true,
			},
			dockerfileExpected: false,
		},
		"default build is ko, ko not exists": {
			args: []string{""},
			cmdExpected: []string{
				"go install github.com/google/ko@latest",
				"KO_DOCKER_REPO=ko.local ko build -B --tags latest",
			},
			lookPathExpected: map[string]bool{
				"ko": false,
			},
			dockerfileExpected: false,
		},
		"ko as builder, specify repo": {
			args: []string{"--repo=ghcr.io/test"},
			cmdExpected: []string{
				"KO_DOCKER_REPO=ghcr.io/test ko build -B --tags latest",
			},
			lookPathExpected: map[string]bool{
				"ko": true,
			},
			dockerfileExpected: false,
		},
		"ko as builder, specify tag": {
			args: []string{"--repo=ghcr.io/test", "--tag=v1"},
			cmdExpected: []string{
				"KO_DOCKER_REPO=ghcr.io/test ko build -B --tags v1",
			},
			lookPathExpected: map[string]bool{
				"ko": true,
			},
			dockerfileExpected: false,
		},
		"docker as builder, docker not exists": {
			args: []string{"--builder=docker"},
			lookPathExpected: map[string]bool{
				"docker": false,
			},
			dockerfileExpected: false,
			expectedError:      "kfn requires that `docker` is installed and on the PATH",
		},
		"docker as builder, docker exists": {
			args: []string{"--builder=docker"},
			cmdExpected: []string{
				"docker build . -f Dockerfile --tag function:latest",
			},
			lookPathExpected: map[string]bool{
				"docker": true,
			},
			dockerfileExpected: true,
		},
		"docker as builder, specify dockerfile": {
			args: []string{"--builder=docker", "--dockerfile=tmp/Dockerfile"},
			cmdExpected: []string{
				"docker build . -f tmp/Dockerfile --tag function:latest",
			},
			lookPathExpected: map[string]bool{
				"docker": true,
			},
			dockerfileExpected: true,
		},
		"docker as builder, specify image": {
			args: []string{"--builder=docker", "--image=dockertest:latest", "--dockerfile=tmp/Dockerfile"},
			cmdExpected: []string{
				"docker build . -f tmp/Dockerfile --tag dockertest:latest",
			},
			lookPathExpected: map[string]bool{
				"docker": true,
			},
			dockerfileExpected: true,
		},
	}
	for name, test := range testcases {
		t.Run(name, func(t *testing.T) {

			r := NewBuildRunner(context.TODO())
			execCmdFn = func(envs []string, name string, args ...string) error {
				fakeExecCmd(t, test.cmdExpected, envs, name, args...)
				return nil
			}
			execLookPathFn = func(file string) (string, error) {
				return "", fakeExecLookPath(t, test.lookPathExpected, file)
			}
			r.Command.SetArgs(test.args)
			err := r.Command.Execute()
			if test.expectedError == "" {
				if err != nil {
					t.Errorf("%v failed. got error: %v", name, err)
				}
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
			if test.dockerfileExpected {
				err = os.Remove("Dockerfile")
				assert.NoError(t, err)
			}
		})
	}
}

func fakeExecCmd(t *testing.T, expectedArgsAndReturns []string, envs []string, name string, args ...string) {
	var c []string
	if name != "go" {
		c = append(c, envs...)
	}
	c = append(c, name)
	c = append(c, args...)
	command := strings.Join(c, " ")
	for _, expected := range expectedArgsAndReturns {
		if expected == command {
			return
		}
	}
	t.Fatalf("unexpected command run %v", command)
}

func fakeExecLookPath(t *testing.T, expectedlookPath map[string]bool, name string) error {
	val, ok := expectedlookPath[name]
	if !ok {
		t.Fatalf("unexpected env var check %v", name)
	}
	if val {
		return nil
	}
	return fmt.Errorf("env var not exists")
}
