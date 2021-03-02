package v0beta1

import (
	skaffoldUtil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/altiscope/platform-stack/pkg/schema/util"

	next "github.com/altiscope/platform-stack/pkg/schema/latest"
)

// Upgrade upgrades a configuration to the next version.
// Config changes from v0beta1 to v0alpha1:
// 1. Additions
//  - Version
// 2. No removal
// 3. No Updates
//  - RequiredVariables is an object, not a list of strings. The keys are the variable names, values are the secret manager id.
func (config *StackConfig) Upgrade() (util.VersionedConfig, error) {
	var newConfig next.StackConfig

	skaffoldUtil.CloneThroughJSON(config.Environments, &newConfig.Environments)
	skaffoldUtil.CloneThroughJSON(config.Stack, &newConfig.Stack)
	ncd := make([]next.ComponentDescription, len(config.Components))
	for i, comp := range config.Components {
		ncd[i].Name = comp.Name
		skaffoldUtil.CloneThroughJSON(comp.Containers, &ncd[i].Containers)
		ncd[i].Exposable = comp.Exposable
		ncd[i].Manifests = comp.Manifests
		ncd[i].TemplateConfig = comp.TemplateConfig
		requiredVariableMap := make(map[string]string)
		for j := 0; j < len(comp.RequiredVariables); j++ {
			requiredVariableMap[comp.RequiredVariables[j]] = comp.RequiredVariables[j]
		}
		ncd[i].RequiredVariables = requiredVariableMap
	}
	skaffoldUtil.CloneThroughJSON(ncd, &newConfig.Components)
	newConfig.ApiVersion = next.Version
	return &newConfig, nil
}
