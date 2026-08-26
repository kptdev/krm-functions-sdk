module github.com/kptdev/krm-functions-sdk/go/fn

go 1.24.0

// We must not include any core k8s APIs (e.g. k8s.io/api) in
// the dependencies, depending on them will likely cause version skew for
// consumers. The dependencies for tests and examples should be isolated.

require (
	github.com/go-errors/errors v1.5.1
	github.com/google/go-cmp v0.7.0
	github.com/kptdev/kpt/api v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.12.0
	go.yaml.in/yaml/v3 v3.0.5
	gotest.tools v2.2.0+incompatible
	k8s.io/klog/v2 v2.140.0
	pgregory.net/rapid v1.3.0
	sigs.k8s.io/kustomize/kyaml v0.21.1
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/sys v0.40.0 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	k8s.io/utils v0.0.0-20260210185600-b8788abfbbc2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
