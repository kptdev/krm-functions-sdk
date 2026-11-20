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
	goerrors "errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kptdev/kpt/api/kubeobject"
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

	registerFlags func(fs *flag.FlagSet) // registers function specific flags
	readInput     func() ([]byte, error) // replaces reading STDIN
	writeOutput   func(out []byte) error // replaces writing STDOUT
}

// WithFlags registers function specific command line flags. register is called
// with a flag.FlagSet before the arguments are parsed. Flags and their values
// are not treated as file paths; "--" ends flag parsing. --help and --doc are
// reserved and handled first, and -h and -help act like --help. Flag values
// are only valid once AsMain has parsed them.
func WithFlags(register func(fs *flag.FlagSet)) Option {
	return func(c *mainConfig) {
		c.registerFlags = register
	}
}

// WithInput replaces reading the ResourceList from STDIN. It is not used in
// file mode.
func WithInput(read func() ([]byte, error)) Option {
	return func(c *mainConfig) {
		c.readInput = read
	}
}

// WithOutput replaces writing the ResourceList to STDOUT. It is also called
// when the function fails, and is not called when there is no output.
func WithOutput(write func(out []byte) error) Option {
	return func(c *mainConfig) {
		c.writeOutput = write
	}
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
// - a `kubeobject.ResourceListProcessor` which implements `Process` method
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
	filePaths, err := positionalArgs(&cfg, os.Args[1:])
	if goerrors.Is(err, flag.ErrHelp) {
		// -h and -help are handled by the flag package, treat them as --help.
		return handleHelp(&cfg)
	}
	if err != nil {
		Logf("failed to parse arguments: %v", err)
		return err
	}

	err = func() error {
		var p kubeobject.ResourceListProcessor
		switch cast := input.(type) {
		case kubeobject.ResourceListProcessor:
			p = cast
		case func(*kubeobject.ResourceList) (bool, error):
			p = kubeobject.ResourceListProcessorFunc(cast)
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
			if outErr := writeResult(&cfg, out); outErr != nil {
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

		in, err := readInput(&cfg)
		if err != nil {
			return err
		}
		out, err := Run(p, in)
		// If there is an error, we don't return the error immediately.
		// We write out to stdout before returning any error.
		if outErr := writeResult(&cfg, out); outErr != nil {
			return outErr
		}
		return err
	}()
	if err != nil {
		Logf("failed to evaluate function: %v", err)
	}
	return err
}

// positionalArgs returns the positional (file path) arguments in args. Without
// function specific flags (WithFlags) every argument starting with "-" is
// skipped. With them, the flags are parsed, and flags and positional
// arguments may be given in any order.
func positionalArgs(cfg *mainConfig, args []string) ([]string, error) {
	var paths []string
	if cfg.registerFlags == nil {
		for _, arg := range args {
			if strings.HasPrefix(arg, "-") {
				continue
			}
			paths = append(paths, arg)
		}
		return paths, nil
	}

	// Everything after the first "--" is a positional argument, even if it
	// starts with "-".
	var afterTerminator []string
	if i := slices.Index(args, "--"); i >= 0 {
		afterTerminator = args[i+1:]
		args = args[:i]
	}

	fs := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	// Errors are reported by AsMain; do not let the flag package print them.
	fs.SetOutput(io.Discard)
	cfg.registerFlags(fs)
	// flag.FlagSet stops parsing at the first positional argument; keep
	// parsing the remainder to allow flags after file paths.
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			return append(paths, afterTerminator...), nil
		}
		paths = append(paths, args[0])
		args = args[1:]
	}
}

// readInput reads the raw function input from the WithInput hook, or STDIN.
func readInput(cfg *mainConfig) ([]byte, error) {
	if cfg.readInput != nil {
		return cfg.readInput()
	}
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("unable to read from stdin: %v", err)
	}
	return in, nil
}

// writeResult writes the raw function output to the WithOutput hook, or STDOUT.
func writeResult(cfg *mainConfig, out []byte) error {
	if cfg.writeOutput != nil {
		// A function that ran always produces a ResourceList. No output means
		// that the input could not be processed; do not call the hook then, so
		// that an existing output file, possibly the input file itself, is not
		// overwritten with nothing.
		if len(out) == 0 {
			return nil
		}
		return cfg.writeOutput(out)
	}
	_, err := os.Stdout.Write(out)
	return err
}

// handleHelp renders help text to STDOUT based on registered docs.
func handleHelp(cfg *mainConfig) error {
	if cfg.readme == nil && cfg.metadata == nil {
		_, err := fmt.Fprint(os.Stdout, "No documentation available. Pass fn.WithDocs to fn.AsMain to enable --help.\n")
		return err
	}

	sections := docs.ParseMarkers(cfg.readme)
	meta, err := docs.ParseMetadata(cfg.metadata)
	if err != nil {
		Logf("warning: invalid metadata YAML: %v", err)
		meta = docs.Metadata{}
	}

	return docs.RenderHelp(os.Stdout, sections, meta)
}

// handleDoc renders JSON documentation to STDOUT based on registered docs.
func handleDoc(cfg *mainConfig) error {
	if cfg.readme == nil && cfg.metadata == nil {
		_, err := fmt.Fprint(os.Stdout, "{}")
		return err
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
func readFilesAsResourceList(paths []string) (*kubeobject.ResourceList, error) {
	rl := &kubeobject.ResourceList{
		FunctionConfig: kubeobject.NewEmptyKubeObject(),
	}
	for _, path := range paths {
		cleanPath := filepath.Clean(path)
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("file %s: %w", path, err)
		}
		// Empty files are valid — proceed with no items from this file.
		if len(strings.TrimSpace(string(data))) == 0 {
			continue
		}
		objects, err := kubeobject.ParseKubeObjects(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse KRM resources from %s: %w", path, err)
		}
		for _, obj := range objects {
			rl.Items = append(rl.Items, obj)
		}
	}
	return rl, nil
}

// Run evaluates the function. input must be a resourceList in yaml format. An
// updated resourceList will be returned.
func Run(p kubeobject.ResourceListProcessor, input []byte) ([]byte, error) {
	if p == nil {
		return nil, fmt.Errorf("nil ResourceListProcessor")
	}
	rl, err := kubeobject.ParseResourceList(input)
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

func Execute(p kubeobject.ResourceListProcessor, r io.Reader, w io.Writer) error {
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

func execute(p kubeobject.ResourceListProcessor, rw *byteReadWriter) error {
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
