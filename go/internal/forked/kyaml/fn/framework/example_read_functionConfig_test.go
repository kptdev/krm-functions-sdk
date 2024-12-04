package framework_test

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/fn/framework"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/fn/framework/command"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/yaml"
)

// In this example, we convert the functionConfig as strong typed object and then
// read a field from the functionConfig object.

func Example_bReadFunctionConfig() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(readFunctionConfig)); err != nil {
		os.Exit(1)
	}
}

func readFunctionConfig(rl *framework.ResourceList) error {
	var sr SetReplicas
	if err := rl.FunctionConfig.As(&sr); err != nil {
		return err
	}
	framework.Logf("desired replicas is %v\n", sr.DesiredReplicas)
	return nil
}

// SetReplicas is the type definition of the functionConfig
type SetReplicas struct {
	yaml.ResourceIdentifier `json:",inline" yaml:",inline"`
	DesiredReplicas         int `json:"desiredReplicas,omitempty" yaml:"desiredReplicas,omitempty"`
}
