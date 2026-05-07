# KRM Functions SDK

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkptdev%2Fkrm-functions-sdk.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fkptdev%2Fkrm-functions-sdk?ref=badge_shield)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/10658/badge)](https://www.bestpractices.dev/projects/10658)

An opinionated Go SDK for implementing [KRM functions](https://kpt.dev/book/05-developing-functions/).

## Quick Start

A KRM function is a program that reads Kubernetes resources from STDIN, transforms or validates them, and writes the result to STDOUT. The SDK handles the I/O — you write the logic.

```go
package main

import (
    "context"
    _ "embed"
    "os"

    "github.com/kptdev/krm-functions-sdk/go/fn"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

type SetLabels struct {
    Labels map[string]string `json:"labels,omitempty"`
}

func (r *SetLabels) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
    for _, obj := range items {
        for k, v := range r.Labels {
            obj.SetLabel(k, v)
        }
    }
    return true
}

func main() {
    runner := fn.WithContext(context.Background(), &SetLabels{})
    if err := fn.AsMain(runner, fn.WithDocs(readme, metadata)); err != nil {
        os.Exit(1)
    }
}
```

A starter template is available at [`go/get-started/`](go/get-started/main.go). For the full walkthrough, see the [Tutorial](docs/tutorial.md).

## How It Works

`fn.AsMain` is the single entrypoint. It handles:

- **STDIN/STDOUT** (default) — reads a ResourceList, processes it, writes the result
- **File mode** — pass file paths as positional args for local debugging
- **`--help`** — prints human-readable docs from embedded README markers
- **`--doc`** — outputs machine-readable JSON (consumed by `kpt fn doc` and catalog pipelines)

Register embedded documentation with `fn.WithDocs`:

```go
fn.AsMain(runner, fn.WithDocs(readme, metadata))
```

The SDK provides two interfaces for implementing functions:

| Interface | Use for | Can add/remove items? | Auto-parses config? |
|---|---|---|---|
| `fn.Runner` | Transformers, validators | No | Yes |
| `fn.ResourceListProcessor` | Generators, complex functions | Yes | No |

See [Interfaces](docs/interfaces.md) for details and code examples.

## Documentation

- [API Reference](https://pkg.go.dev/github.com/kptdev/krm-functions-sdk/go/fn) — Go API docs
- [Tutorial](docs/tutorial.md) — end-to-end function development
- [Interfaces](docs/interfaces.md) — Runner vs ResourceListProcessor
- [Testing](docs/testing.md) — golden test patterns
- [Containerizing](docs/containerizing.md) — building and running function images

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on DCO sign-off, copyright headers, and code review process.

## Issues

Please [open issues](https://github.com/kptdev/kpt/issues) at [kptdev/kpt](https://github.com/kptdev/kpt/).

## License

Code is under the [Apache License 2.0](LICENSE), documentation is [CC BY 4.0](LICENSE-documentation).

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkptdev%2Fkrm-functions-sdk.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fkptdev%2Fkrm-functions-sdk?ref=badge_large)

## Governance

The governance of the kpt project is described in the
[governance repo](https://github.com/kptdev/governance).

## Code of Conduct

The kpt project follows the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md).
More information is [here](code-of-conduct.md).

## CNCF

The kpt project is a [CNCF Sandbox](https://www.cncf.io/sandbox-projects/) project.
