module github.com/kptdev/krm-functions-sdk/go/fn

go 1.24.0

require (
	github.com/go-errors/errors v1.5.1
	github.com/google/go-cmp v0.7.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.11.1
	go.yaml.in/yaml/v3 v3.0.4
	gotest.tools v2.2.0+incompatible
	// We must not include any core k8s APIs (e.g. k8s.io/api) in
	// the dependencies, depending on them will likely to cause version skew for
	// consumers. The dependencies for tests and examples should be isolated.
	k8s.io/klog/v2 v2.140.0
	pgregory.net/rapid v1.3.0
	sigs.k8s.io/kustomize/kyaml v0.21.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
)
