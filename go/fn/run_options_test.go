// Copyright 2026 The kpt Authors
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
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testResourceList = `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: cm
`

// testFlags holds the values of the flags registered by registerTestFlags.
type testFlags struct {
	input  string
	output string
	json   bool
}

func (f *testFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&f.input, "i", "", "input")
	fs.StringVar(&f.input, "input", "", "input")
	fs.StringVar(&f.output, "o", "", "output")
	fs.StringVar(&f.output, "output", "", "output")
	fs.BoolVar(&f.json, "json", false, "json")
}

// recordingProcessor records how many items it saw.
type recordingProcessor struct {
	items int
	calls int
	err   error
	fail  bool
}

func (p *recordingProcessor) Process(rl *kubeobject.ResourceList) (bool, error) {
	p.calls++
	p.items = len(rl.Items)
	return !p.fail, p.err
}

func TestAsMain_WithFlags_ParsesFlagsAndKeepsPositionals(t *testing.T) {
	dir := t.TempDir()
	file1 := filepath.Join(dir, "a.yaml")
	file2 := filepath.Join(dir, "b.yaml")
	require.NoError(t, os.WriteFile(file1, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: a\n"), 0o600))
	require.NoError(t, os.WriteFile(file2, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: b\n"), 0o600))

	// Flags, values and positionals in mixed order. Values of flags must not
	// be treated as file paths.
	setArgs(t, []string{"cmd", "-i", "ignored-in.yaml", file1, "--json", "--output=ignored-out.yaml", file2})

	var f testFlags
	p := &recordingProcessor{}
	output := captureStdout(t, func() {
		require.NoError(t, AsMain(p, WithFlags(f.register)))
	})

	assert.Equal(t, 2, p.items, "only the two real file paths must be read")
	assert.Equal(t, "ignored-in.yaml", f.input)
	assert.Equal(t, "ignored-out.yaml", f.output)
	assert.True(t, f.json)
	assert.Contains(t, output, "name: a")
	assert.Contains(t, output, "name: b")
}

func TestAsMain_WithoutFlags_SkipsDashArguments(t *testing.T) {
	// Existing behavior: any argument starting with "-" is skipped.
	dir := t.TempDir()
	file := filepath.Join(dir, "a.yaml")
	require.NoError(t, os.WriteFile(file, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: a\n"), 0o600))
	setArgs(t, []string{"cmd", "--unknown", file})

	p := &recordingProcessor{}
	captureStdout(t, func() {
		require.NoError(t, AsMain(p))
	})
	assert.Equal(t, 1, p.items)
}

func TestAsMain_WithFlags_UnknownFlagIsError(t *testing.T) {
	setArgs(t, []string{"cmd", "--nope"})

	var f testFlags
	p := &recordingProcessor{}
	var err error
	captureStderr(t, func() {
		err = AsMain(p, WithFlags(f.register))
	})
	require.Error(t, err)
	assert.Zero(t, p.calls, "function must not run when flag parsing fails")
}

func TestAsMain_WithFlags_HelpAndDocTakePrecedence(t *testing.T) {
	for _, flagName := range []string{"--help", "--doc"} {
		setArgs(t, []string{"cmd", "-i", "in.yaml", flagName, "--json"})

		var f testFlags
		p := &recordingProcessor{}
		output := captureStdout(t, func() {
			require.NoError(t, AsMain(p, WithFlags(f.register),
				WithDocs([]byte("<!--mdtogo:Short-->\nShort text\n<!--mdtogo-->\n"), []byte("image: test\n"))))
		})
		assert.Contains(t, output, "Short text", flagName)
		assert.Zero(t, p.calls, flagName)
		assert.Empty(t, f.input, "flags must not be parsed when %s is handled", flagName)
	}
}

func TestAsMain_WithInputWithOutput_StdinMode(t *testing.T) {
	setArgs(t, []string{"cmd", "--json"})

	var f testFlags
	var written []byte
	p := &recordingProcessor{}
	output := captureStdout(t, func() {
		require.NoError(t, AsMain(p,
			WithFlags(f.register),
			WithInput(func() ([]byte, error) {
				assert.True(t, f.json, "flags must be parsed before the input hook runs")
				return []byte(testResourceList), nil
			}),
			WithOutput(func(out []byte) error {
				written = out
				return nil
			}),
		))
	})

	assert.Empty(t, output, "nothing must be written to STDOUT when WithOutput is used")
	assert.Equal(t, 1, p.items)
	assert.Contains(t, string(written), "kind: ConfigMap")
}

func TestAsMain_WithOutput_WrittenBeforeFunctionErrorIsReturned(t *testing.T) {
	setArgs(t, []string{"cmd"})

	fnErr := errors.New("validation failed")
	var written []byte
	err := AsMain(&recordingProcessor{err: fnErr},
		WithInput(func() ([]byte, error) { return []byte(testResourceList), nil }),
		WithOutput(func(out []byte) error { written = out; return nil }),
	)

	assert.ErrorIs(t, err, fnErr)
	assert.Contains(t, string(written), "kind: ConfigMap", "output must be written even when the function fails")
}

func TestAsMain_WithInput_ErrorIsReturnedAndFunctionNotRun(t *testing.T) {
	setArgs(t, []string{"cmd"})

	inErr := errors.New("cannot read input")
	p := &recordingProcessor{}
	var err error
	captureStderr(t, func() {
		err = AsMain(p, WithInput(func() ([]byte, error) { return nil, inErr }))
	})
	assert.ErrorIs(t, err, inErr)
	assert.Zero(t, p.calls)
}

func TestAsMain_WithOutput_ErrorIsReturned(t *testing.T) {
	setArgs(t, []string{"cmd"})

	outErr := errors.New("cannot write output")
	var err error
	captureStderr(t, func() {
		err = AsMain(&recordingProcessor{},
			WithInput(func() ([]byte, error) { return []byte(testResourceList), nil }),
			WithOutput(func([]byte) error { return outErr }),
		)
	})
	assert.ErrorIs(t, err, outErr)
}

func TestAsMain_WithOutput_FileMode(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.yaml")
	require.NoError(t, os.WriteFile(file, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: a\n"), 0o600))
	setArgs(t, []string{"cmd", file})

	var written []byte
	output := captureStdout(t, func() {
		require.NoError(t, AsMain(&recordingProcessor{},
			WithOutput(func(out []byte) error { written = out; return nil })))
	})
	assert.Empty(t, output)
	assert.Contains(t, string(written), "name: a")
}

func TestAsMain_WithFlags_TerminatorEndsFlagParsing(t *testing.T) {
	dir := t.TempDir()
	// A file name that looks like a flag is only usable after "--".
	dashFile := filepath.Join(dir, "-dash.yaml")
	plainFile := filepath.Join(dir, "plain.yaml")
	require.NoError(t, os.WriteFile(dashFile, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: dash\n"), 0o600))
	require.NoError(t, os.WriteFile(plainFile, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: plain\n"), 0o600))

	// Run from dir so that the relative file name starts with "-".
	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { assert.NoError(t, os.Chdir(origWd)) })

	setArgs(t, []string{"cmd", "--json", "plain.yaml", "--", "-dash.yaml", "--json"})

	var f testFlags
	p := &recordingProcessor{}
	var err2 error
	captureStderr(t, func() {
		captureStdout(t, func() {
			err2 = AsMain(p, WithFlags(f.register))
		})
	})

	// "--json" after "--" is a file path, which does not exist, so reading fails.
	require.Error(t, err2)
	assert.Contains(t, err2.Error(), "--json")
	assert.True(t, f.json, "flags before the terminator must still be parsed")
	assert.Zero(t, p.calls)
}

func TestAsMain_WithFlags_TerminatorKeepsDashPrefixedFilePaths(t *testing.T) {
	dir := t.TempDir()
	dashFile := filepath.Join(dir, "-dash.yaml")
	require.NoError(t, os.WriteFile(dashFile, []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: dash\n"), 0o600))
	origWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { assert.NoError(t, os.Chdir(origWd)) })

	setArgs(t, []string{"cmd", "--json", "--", "-dash.yaml"})

	var f testFlags
	p := &recordingProcessor{}
	output := captureStdout(t, func() {
		require.NoError(t, AsMain(p, WithFlags(f.register)))
	})
	assert.True(t, f.json)
	assert.Equal(t, 1, p.items)
	assert.Contains(t, output, "name: dash")
}

func TestAsMain_WithFlags_SingleDashIsPositional(t *testing.T) {
	setArgs(t, []string{"cmd", "-"})

	var f testFlags
	p := &recordingProcessor{}
	var err error
	captureStderr(t, func() {
		err = AsMain(p, WithFlags(f.register))
	})
	// "-" is a positional argument (a file path), not a flag.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "-")
	assert.Zero(t, p.calls)
}

func TestAsMain_WithFlags_FlagMissingValueIsError(t *testing.T) {
	setArgs(t, []string{"cmd", "-i"})

	var f testFlags
	p := &recordingProcessor{}
	var err error
	captureStderr(t, func() {
		err = AsMain(p, WithFlags(f.register))
	})
	require.Error(t, err)
	assert.Zero(t, p.calls)
}

func TestAsMain_WithOutput_NotCalledWhenInputCannotBeParsed(t *testing.T) {
	inputs := map[string]string{
		"not a ResourceList": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: x\n",
		"broken yaml":        "items: [unclosed\n",
	}
	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			setArgs(t, []string{"cmd"})

			p := &recordingProcessor{}
			called := false
			var err error
			captureStderr(t, func() {
				err = AsMain(p,
					WithInput(func() ([]byte, error) { return []byte(input), nil }),
					WithOutput(func([]byte) error { called = true; return nil }),
				)
			})

			require.Error(t, err)
			assert.Zero(t, p.calls, "function must not run on unparseable input")
			assert.False(t, called, "output must not be written, an existing output file would be overwritten with nothing")
		})
	}
}

func TestAsMain_WithOutput_InPlaceFileIsKeptOnInvalidInput(t *testing.T) {
	// The input and output flags can point to the same file. An invalid input
	// must not destroy the file.
	dir := t.TempDir()
	path := filepath.Join(dir, "resources.yaml")
	content := []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: x\n")
	require.NoError(t, os.WriteFile(path, content, 0o600))
	setArgs(t, []string{"cmd", "--input", path, "--output", path})

	var f testFlags
	var err error
	captureStderr(t, func() {
		err = AsMain(&recordingProcessor{},
			WithFlags(f.register),
			WithInput(func() ([]byte, error) { return os.ReadFile(f.input) }),
			WithOutput(func(out []byte) error { return os.WriteFile(f.output, out, 0o600) }),
		)
	})

	require.Error(t, err)
	got, readErr := os.ReadFile(filepath.Clean(path))
	require.NoError(t, readErr)
	assert.Equal(t, content, got, "the file must be unchanged")
}

func TestAsMain_WithFlags_ShortHelpFlagsAreHelp(t *testing.T) {
	readme := []byte("<!--mdtogo:Short-->\nShort text\n<!--mdtogo-->\n")
	for _, helpFlag := range []string{"-h", "-help"} {
		t.Run(helpFlag, func(t *testing.T) {
			setArgs(t, []string{"cmd", "--json", helpFlag})

			var f testFlags
			p := &recordingProcessor{}
			var stdout string
			stderr := captureStderr(t, func() {
				stdout = captureStdout(t, func() {
					require.NoError(t, AsMain(p, WithFlags(f.register), WithDocs(readme, []byte("image: test\n"))))
				})
			})

			assert.Contains(t, stdout, "Short text")
			assert.Empty(t, stderr, "the flag package must not print its own usage")
			assert.Zero(t, p.calls)
		})
	}
}

func TestAsMain_WithFlags_ShortHelpFlagWithoutDocs(t *testing.T) {
	setArgs(t, []string{"cmd", "-h"})

	var f testFlags
	output := captureStdout(t, func() {
		require.NoError(t, AsMain(&recordingProcessor{}, WithFlags(f.register)))
	})
	assert.Contains(t, output, "No documentation available")
}

func TestAsMain_WithFlags_FunctionMayDefineItsOwnShortH(t *testing.T) {
	setArgs(t, []string{"cmd", "-h", "value"})

	var h string
	p := &recordingProcessor{}
	require.NoError(t, AsMain(p,
		WithFlags(func(fs *flag.FlagSet) { fs.StringVar(&h, "h", "", "a flag named h") }),
		WithInput(func() ([]byte, error) { return []byte(testResourceList), nil }),
		WithOutput(func([]byte) error { return nil }),
	))
	assert.Equal(t, "value", h)
	assert.Equal(t, 1, p.calls, "the function must run, -h is its own flag")
}

func TestAsMain_WithFlags_ErrorsAreNotPrintedByTheFlagPackage(t *testing.T) {
	setArgs(t, []string{"cmd", "--nope"})

	var f testFlags
	var err error
	stderr := captureStderr(t, func() {
		err = AsMain(&recordingProcessor{}, WithFlags(f.register))
	})
	require.Error(t, err)
	// The error is reported once, by AsMain, not also as a flag package usage dump.
	assert.Equal(t, 1, strings.Count(stderr, "not defined"), stderr)
}
