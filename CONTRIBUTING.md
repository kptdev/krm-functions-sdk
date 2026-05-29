# Contributing to krm-functions-sdk

We would love to accept your contributions to this project. There are just a few
small guidelines you need to follow.

## Developer Certificate of Origin (DCO)

Contributors to this project should state that they agree with the terms published
at https://developercertificate.org/ for their contribution. To do this when
creating a commit with the Git CLI, a sign-off can be added with
[the -s option](https://git-scm.com/docs/git-commit#git-commit--s). The sign-off
is stored as part of the commit message itself.

## Copyright notices

All files should have the copyright notice.
```
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
```

If the file has never been modified: use the creation year only.

* Example: `Copyright 2026 The kpt Authors`

If the file has been modified: use a year range from creation to last modification.

* Example: `Copyright 2024-2026 The kpt Authors`

## Building and Testing

The SDK uses a Makefile-based workflow. From the repository root:

```bash
# Run all checks (fix, vet, fmt, test, lint)
make go

# Run only tests
cd go && make test

# Run only linting
cd go && make lint

# Tidy all go.mod files
make tidy
```

The CI script (`hack/ci-validate-go.sh`) runs `make go` and then checks that no
files have been modified. If the CI script fails with "files are not to date", run `make go`
locally and commit the changes.

## Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult [GitHub Help] for more
information on using pull requests.

Process for code reviews. Before requesting human review, a PR must:

* All tests passing
* All linting passing
* Meeting project code quality requirements, including passing all configured
  static analysis and not reducing automated test coverage
* The comments from the first run of automatically generated comments (AI
  generated comments, bot generated comments, etc.) of the PR are addressed
  (addressing further re-runs of AI are optional)
* If it is not possible to resolve an automatic comment, please add a sub-comment
  indicating why the automated comment cannot be resolved or ask for help in
  resolving the comment
* The PR description states whether AI was used to help create the PR; if so, it
  lists the AI tools used and the areas where they were used

## Declare any use of AI

> In addition to the above, the use of AI in the creation of PRs is allowed, but
> you must declare any use of AI and you must be able to explain the PR code
> independently of any AI tools.

Update the PR description to state whether you used AI to help you create this
PR; if so, list the AI tools you have used and in what areas.

For example:
```text
I have used AI in the creation of this PR.

I have used the following AI tools:
- GitHub Copilot to analyse the code
- Kiro to generate the implementation and tests
```

### Attribute AI in the Git commit messages

Following the [guidance of the Linux kernel](https://docs.kernel.org/process/coding-assistants.html#attribution)
we recommend the attribution of AI tools in the commit messages using the following format:

```text
Assisted-by: AGENT_NAME:MODEL_VERSION [TOOL1] [TOOL2]
```

## Community Guidelines

This project follows a [Code of Conduct].

## Community Discussion Groups

1. Join our [Slack channel](https://kubernetes.slack.com/channels/kpt)
1. Join our [Discussions](https://github.com/kptdev/kpt/discussions)

## Governance

The governance of the kpt project is described in the
[governance repo](https://github.com/kptdev/governance).

[GitHub Help]: https://help.github.com/articles/about-pull-requests/
[Code of Conduct]: code-of-conduct.md
