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
// 3. Updates
//  - RequiredVariables is an object, not a list of strings. The keys are the variable names, values are the secret manager id.
func (config *StackConfig) Upgrade() (util.VersionedConfig, error) {
	var newConfig next.StackConfig
	var newComps []next.ComponentDescription
	skaffoldUtil.CloneThroughYAML(config.Components, &newComps)
	var newEnvs []next.EnvironmentDescription
	skaffoldUtil.CloneThroughYAML(config.Environments, &newEnvs)
	var newStack next.StackDescription
	skaffoldUtil.CloneThroughYAML(config.Stack, &newConfig.Stack)
	nextConfig := &next.StackConfig{
		ApiVersion:   next.Version,
		Components:   newComps,
		Environments: newEnvs,
		Stack:        newStack,
	}
	return nextConfig, nil
}
