// Copyright 2022-2026 The kpt Authors
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
	"io"
	"os"
	"slices"
	"strings"

	"github.com/kptdev/krm-functions-sdk/go/fn/internal/docs"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// Option configures fn.AsMain behavior.
type Option func(*mainConfig)

// mainConfig holds configuration gathered from Options.
type mainConfig struct {
	readme   []byte // raw embedded README.md content
	metadata []byte // raw embedded metadata.yaml content
}

// WithDocs registers embedded README and metadata content for --help and --doc.
func WithDocs(readme []byte, meta []byte) Option {
	return func(c *mainConfig) {
		c.readme = readme
		c.metadata = meta
	}
}

// AsMain evaluates a KRM function. By default it reads a ResourceList from
// STDIN, processes it, and writes the result to STDOUT.
//
// `input` can be
// - a `ResourceListProcessor` which implements `Process` method
// - a function `Runner` which implements `Run` method
//
// Invocation modes (checked in this order):
//   - --help: prints human-readable documentation to STDOUT and returns nil.
//   - --doc: prints machine-readable JSON documentation to STDOUT and returns nil.
//   - positional file args: reads KRM resources from files instead of STDIN.
//   - no args: reads ResourceList from STDIN (default behavior).
//
// Options configure additional behavior such as documentation support
// via WithDocs. Existing callers with no options continue to work unchanged.
func AsMain(input any, opts ...Option) error {
	// Apply options to build configuration.
	var cfg mainConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	// Check for --help and --doc flags before reading STDIN.
	// --help always takes precedence over --doc regardless of argument order.
	if slices.Contains(os.Args[1:], "--help") {
		return handleHelp(&cfg)
	}
	if slices.Contains(os.Args[1:], "--doc") {
		return handleDoc(&cfg)
	}

	// Collect non-flag positional arguments (file paths).
	// Skip any argument that looks like a flag (starts with "-" or "--").
	var filePaths []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		filePaths = append(filePaths, arg)
	}

	err := func() error {
		var p ResourceListProcessor
		switch input := input.(type) {
		case runnerProcessor:
			p = input
		case ResourceListProcessorFunc:
			p = input
		default:
			return fmt.Errorf("unknown input type %T", input)
		}

		// If file paths are provided, use file mode instead of STDIN.
		if len(filePaths) > 0 {
			rl, err := readFilesAsResourceList(filePaths)
			if err != nil {
				return err
			}
			success, fnErr := p.Process(rl)
			out, yamlErr := rl.ToYAML()
			if yamlErr != nil {
				return yamlErr
			}
			_, outErr := os.Stdout.Write(out)
			if outErr != nil {
				return outErr
			}
			if fnErr != nil {
				return fnErr
			}
			if !success {
				return fmt.Errorf("error: function failure")
			}
			return nil
		}

		in, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("unable to read from stdin: %v", err)
		}
		out, err := Run(p, in)
		// If there is an error, we don't return the error immediately.
		// We write out to stdout before returning any error.
		_, outErr := os.Stdout.Write(out)
		if outErr != nil {
			return outErr
		}
		return err
	}()
	if err != nil {
		Logf("failed to evaluate function: %v", err)
	}
	return err
}

// handleHelp renders help text to STDOUT based on registered docs.
func handleHelp(cfg *mainConfig) error {
	if cfg.readme == nil && cfg.metadata == nil {
		fmt.Fprint(os.Stdout, "No documentation available. Pass fn.WithDocs to fn.AsMain to enable --help.\n")
		return nil
	}

	sections := docs.ParseMarkers(cfg.readme)
	meta, err := docs.ParseMetadata(cfg.metadata)
	if err != nil {
		Logf("warning: invalid metadata YAML: %v", err)
		meta = docs.Metadata{}
	}

	docs.RenderHelp(os.Stdout, sections, meta)
	return nil
}

// handleDoc renders JSON documentation to STDOUT based on registered docs.
func handleDoc(cfg *mainConfig) error {
	if cfg.readme == nil && cfg.metadata == nil {
		fmt.Fprint(os.Stdout, "{}")
		return nil
	}

	sections := docs.ParseMarkers(cfg.readme)
	meta, err := docs.ParseMetadata(cfg.metadata)
	if err != nil {
		Logf("warning: invalid metadata YAML: %v", err)
		meta = docs.Metadata{}
	}

	return docs.RenderDoc(os.Stdout, sections, meta)
}

// readFilesAsResourceList reads KRM YAML from the given file paths,
// assembles them into a ResourceList with an empty FunctionConfig.
// Each file is parsed as one or more KRM YAML documents (separated by ---).
// Empty files are valid (no items added). Returns a descriptive error if a
// file does not exist or contains invalid YAML.
func readFilesAsResourceList(paths []string) (*ResourceList, error) {
	rl := &ResourceList{
		FunctionConfig: NewEmptyKubeObject(),
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file not found: %s", path)
			}
			return nil, fmt.Errorf("failed to read file %s: %v", path, err)
		}
		// Empty files are valid — proceed with no items from this file.
		if len(strings.TrimSpace(string(data))) == 0 {
			continue
		}
		objects, err := ParseKubeObjects(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse KRM resources from %s: %v", path, err)
		}
		for _, obj := range objects {
			rl.Items = append(rl.Items, obj)
		}
	}
	return rl, nil
}

// Run evaluates the function. input must be a resourceList in yaml format. An
// updated resourceList will be returned.
func Run(p ResourceListProcessor, input []byte) ([]byte, error) {
	switch input := p.(type) {
	case runnerProcessor:
		p = input
	case ResourceListProcessorFunc:
		p = input
	default:
		return nil, fmt.Errorf("unknown input type %T", input)
	}
	rl, err := ParseResourceList(input)
	if err != nil {
		return nil, err
	}
	success, fnErr := p.Process(rl)
	out, yamlErr := rl.ToYAML()
	if yamlErr != nil {
		return out, yamlErr
	}
	if fnErr != nil {
		return out, fnErr
	}
	if !success {
		return out, fmt.Errorf("error: function failure")
	}
	return out, nil
}

func Execute(p ResourceListProcessor, r io.Reader, w io.Writer) error {
	rw := &byteReadWriter{
		kio.ByteReadWriter{
			Reader: r,
			Writer: w,
			// We should not set the id annotation in the function, since we should not
			// overwrite what the orchestrator set.
			OmitReaderAnnotations: true,
			// We should not remove the id annotations in the function, since the
			// orchestrator (e.g. kpt) may need them.
			KeepReaderAnnotations: true,
		},
	}
	return execute(p, rw)
}

func execute(p ResourceListProcessor, rw *byteReadWriter) error {
	// Read the input
	rl, err := rw.Read()
	if err != nil {
		return errors.WrapPrefixf(err, "failed to read ResourceList input")
	}
	success, fnErr := p.Process(rl)
	// Write the output
	if err := rw.Write(rl); err != nil {
		return errors.WrapPrefixf(err, "failed to write ResourceList output")
	}
	if fnErr != nil {
		return fnErr
	}
	if !success {
		return fmt.Errorf("error: function failure")
	}
	return nil
}
