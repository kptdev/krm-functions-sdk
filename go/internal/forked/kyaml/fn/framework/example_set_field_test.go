package framework_test

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/fn/framework"
	"github.com/GoogleContainerTools/kpt-functions-sdk/krmfn/internal/forked/kyaml/fn/framework/command"
)

// In this example, we read a field from the input object and print it to the log.

func Example_cSetField() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(setField)); err != nil {
		os.Exit(1)
	}
}

func setField(rl *framework.ResourceList) error {
	for _, obj := range rl.Items {
		if obj.GetKind() == "Deployment" && obj.GetName() == "nginx" {
			return obj.Set(10, "spec", "replicas")
		}
	}
	return nil
}
