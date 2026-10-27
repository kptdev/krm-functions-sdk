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

package testhelpers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kptdev/kpt/api/kubeobject"
)

// RunGoldenTests provides the test infra to run golden test.
// "basedir" should be the parent directory, under where the sub-directories contains test data.
// For example, the "testdata" is the basedir. It contains two cases "test1" and "test2"
//
//	└── testdata
//	    └── test1
//	        ├── _expected.yaml
//	        ├── _fnconfig.yaml
//	        └── resources.yaml
//	    └── test2
//	        ├── _expected.yaml
//	        ├── _fnconfig.yaml
//	        └── resources.yaml
//
// - "krmFunction" should be your ResourceListProcessor implementation.
func RunGoldenTests(t *testing.T, basedir string, krmFunction kubeobject.ResourceListProcessor) {
	dirEntries, err := os.ReadDir(basedir)
	if err != nil {
		t.Fatalf("ReadDir(%q) failed: %v", basedir, err)
	}

	for _, dirEntry := range dirEntries {
		dir := filepath.Join(basedir, dirEntry.Name())
		if !dirEntry.IsDir() {
			t.Errorf("expected directory, found %s", dir)
			continue
		}

		t.Run(dir, func(t *testing.T) {
			files, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("failed to read directory %q: %v", basedir, err)
			}
			sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
			var items []*kubeobject.KubeObject
			for _, f := range files {
				if strings.HasPrefix(f.Name(), "_") {
					continue
				}
				// Users can put other types of files to the test dir, but they won't be read.
				// A default kpt package contains README file.
				if !IsValidYAMLOrKptfile(f.Name()) {
					continue
				}
				fileItems := mustParseFile(t, filepath.Join(dir, f.Name()))
				items = append(items, fileItems...)
			}

			config := mustParseFile(t, filepath.Join(dir, "_fnconfig.yaml"))

			var functionConfig *kubeobject.KubeObject
			if len(config) == 0 {
				functionConfig = kubeobject.NewEmptyKubeObject()
			} else if len(config) == 1 {
				functionConfig = config[0]
			} else {
				t.Fatalf("found multiple config objects in %s", filepath.Join(dir, "_fnconfig.yaml"))
			}
			rl := &kubeobject.ResourceList{Items: items, FunctionConfig: functionConfig}
			_, err = krmFunction.Process(rl)
			if err != nil {
				t.Fatalf("run failed unexpectedly: %v", err)
			}
			rlYAML, err := rl.ToYAML()
			if err != nil {
				t.Fatalf("failed to convert resource list to yaml: %v", err)
			}
			p := filepath.Join(dir, "_expected.yaml")
			CompareGoldenFile(t, p, rlYAML)
		})
	}
}

func IsValidYAMLOrKptfile(fname string) bool {
	if strings.HasSuffix(fname, ".yaml") {
		return true
	}
	if strings.HasSuffix(fname, ".yml") {
		return true
	}
	if fname == "Kptfile" {
		return true
	}
	return false
}

// MustReadFile reads the data from "expectedPath"
func MustReadFile(t *testing.T, expectedPath string) []byte {
	b, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read file %q: %v", expectedPath, err)
	}
	return b
}

// CompareGoldenFile compares the "got" data with the data stored in a "expectedPath".
func CompareGoldenFile(t *testing.T, expectedPath string, got []byte) {
	if os.Getenv("WRITE_GOLDEN_OUTPUT") != "" {
		// Short-circuit when the output is correct
		b, err := os.ReadFile(expectedPath)
		if err == nil && bytes.Equal(b, got) {
			return
		}
		if err = os.WriteFile(expectedPath, got, 0600); err != nil {
			t.Fatalf("failed to write golden output %s: %v", expectedPath, err)
		}
		t.Errorf("wrote output to %s", expectedPath)
	} else {
		want := MustReadFile(t, expectedPath)
		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			t.Errorf("unexpected diff in %s: %s", expectedPath, diff)
		}
	}
}

// CopyDir copies an entire directory from "src" to "dest"
func CopyDir(src, dest string) error {
	srcFiles, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("ReadDir(%q) failed: %w", src, err)
	}
	for _, srcFile := range srcFiles {
		srcPath := filepath.Join(src, srcFile.Name())
		destPath := filepath.Join(dest, srcFile.Name())
		if srcFile.IsDir() {
			if err = CopyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyFile copies a single file from "src" to "dest"
func CopyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("OpenFile(%q) failed: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create(%q) failed: %w", dest, err)
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("byte copy from %s to %s failed: %w", src, dest, err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("close(%q) failed: %w", dest, err)
	}

	return nil
}

// CompareDir compares the contents of two directories
// only compare KRM YAML resources?
func CompareDir(t *testing.T, expectDir, actualDir string) {
	expectFiles, err := os.ReadDir(expectDir)
	if err != nil {
		t.Fatalf("failed to read expectation directory %q: %v", expectDir, err)
	}
	expectFileMap := make(map[string]os.DirEntry)
	for _, expectFile := range expectFiles {
		expectFileMap[expectFile.Name()] = expectFile
	}

	actualFiles, err := os.ReadDir(actualDir)
	if err != nil {
		t.Fatalf("failed to read actual directory %q: %v", actualDir, err)
	}
	actualFileMap := make(map[string]os.DirEntry)
	for _, actualFile := range actualFiles {
		actualFileMap[actualFile.Name()] = actualFile
	}

	for _, expectFile := range expectFiles {
		name := expectFile.Name()
		actualFile := actualFileMap[name]
		if actualFile == nil {
			t.Errorf("expected file %s not found", name)
			continue
		}

		if expectFile.IsDir() {
			if !actualFile.IsDir() {
				t.Errorf("expected file %s was not a directory", name)
				continue
			}
			CompareDir(t, filepath.Join(expectDir, name), filepath.Join(actualDir, name))
		} else {
			if actualFile.IsDir() {
				t.Errorf("expected file %s was not a file", name)
				continue
			}
			CompareFile(t, expectDir, actualDir, name)
		}
	}

	for _, actualFile := range actualFiles {
		name := actualFile.Name()
		expectFile := expectFileMap[name]
		if expectFile == nil {
			t.Errorf("additional file %s found in output", name)
			continue
		}
	}
}

// CompareFile compares a single file of the same relative path between "expectDir" and "actualDir"
func CompareFile(t *testing.T, expectDir, actualDir string, relPath string) {
	expectAbs := filepath.Join(expectDir, relPath)

	actualAbs := filepath.Join(actualDir, relPath)
	actualBytes, err := os.ReadFile(actualAbs)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("expected file %s not found", relPath)
		} else {
			t.Fatalf("error reading file %s: %v", actualAbs, err)
		}
	}

	CompareGoldenFile(t, expectAbs, actualBytes)
}

func mustParseFile(t *testing.T, path string) kubeobject.KubeObjects {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	b := MustReadFile(t, path)

	objects, err := kubeobject.ParseKubeObjects(b)
	if err != nil {
		t.Fatalf("failed to parse objects from file %q: %v", path, err)
	}
	return objects
}
