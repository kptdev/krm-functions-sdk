# hello-world

Note: Please ensure you follow the [kpt doc style guide].

## Overview

<!--mdtogo:Short-->

A "hello world" example KRM function built with the Go SDK.

<!--mdtogo-->

This is the get-started example for the KRM Functions Go SDK. It is a small,
working mutator function that builds a greeting from its functionConfig and
stamps it as an annotation (`example.kpt.dev/greeting`) on every resource it
receives. Copy this directory as the starting point for your own function.

[//]: <> (Note: The content between `<!--mdtogo:Short-->` and the following
`<!--mdtogo-->` will be used as the short description for the command.)

<!--mdtogo:Long-->

## Usage

The function reads a `HelloWorld` functionConfig, forms the message
`"<greeting>, <name>!"`, and sets it as the `example.kpt.dev/greeting`
annotation on every resource in the package. It runs as a mutator in a `kpt`
pipeline or standalone (STDIN/STDOUT, file mode, `--help`, `--doc`).

### FunctionConfig

The functionConfig `kind` matches the Go struct name (`HelloWorld`). Each field
is populated from the matching key via its JSON tag.

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: HelloWorld
metadata:
  name: my-config
greeting: Hello
name: world
```

- `greeting` (optional): the salutation. Defaults to `Hello`.
- `name` (optional): who to greet. Defaults to `world`.

[//]: <> (Note: The content between `<!--mdtogo:Long-->` and the following
`<!--mdtogo-->` will be used as the long description for the command.)

<!--mdtogo-->

## Examples

<!--mdtogo:Examples-->

Greet every resource with the default `Hello, world!`:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: HelloWorld
```

Customize the greeting:

```yaml
apiVersion: fn.kpt.dev/v1alpha1
kind: HelloWorld
greeting: Howdy
name: partner
```

[//]: <> (Note: The content between `<!--mdtogo:Examples-->` and the following
`<!--mdtogo-->` will be used as the examples for the command.)

<!--mdtogo-->

[kpt doc style guide]: https://github.com/kptdev/kpt/blob/main/docs/style-guides/docs.md
