module github.com/GoogleContainerTools/kpt-functions-sdk/go/fn

go 1.18

require (
	github.com/GoogleContainerTools/kpt-functions-sdk/go/api v0.0.0-20220720212527-133180134b93
	github.com/go-errors/errors v1.0.1
	github.com/google/go-cmp v0.5.9
	github.com/stretchr/testify v1.8.0
	google.golang.org/genproto/googleapis/api v0.0.0-20230803162519-f966b187b2e5 // indirect
	k8s.io/apimachinery v0.24.0
	// We must not include any core k8s APIs (e.g. k8s.io/api) in
	// the dependencies, depending on them will likely to cause version skew for
	// consumers. The dependencies for tests and examples should be isolated.
	k8s.io/klog/v2 v2.60.1
	sigs.k8s.io/kustomize/kyaml v0.13.7-0.20220418212550-9d5491c2e20c

)

require github.com/google/cel-go v0.18.0

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230305170008-8188dc5388df // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20220722155223-a9213eeb770e // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230803162519-f966b187b2e5 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220401212409-b28bf2818661 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
