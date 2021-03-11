package v0beta1

import (
	skaffoldUtil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/altiscope/platform-stack/pkg/schema/util"

	next "github.com/altiscope/platform-stack/pkg/schema/latest"
)

// Upgrade upgrades a configuration to the next version.
// 1. Additions
//  - Version
//  - Environments list added to ComponentDescription
// 2. No removal
// 3. No Updates
func (config *StackConfig) Upgrade() (util.VersionedConfig, error) {
	var newComps []next.ComponentDescription
	skaffoldUtil.CloneThroughYAML(config.Components, &newComps)
	for i, _ := range newComps {
		newComps[i].Environments = []string{}
	}
	var newEnvs []next.EnvironmentDescription
	skaffoldUtil.CloneThroughYAML(config.Environments, &newEnvs)
	var newStack next.StackDescription
	skaffoldUtil.CloneThroughYAML(config.Stack, &newStack)
	nextConfig := &next.StackConfig{
		ApiVersion:   next.Version,
		Components:   newComps,
		Environments: newEnvs,
		Stack:        newStack,
	}
	return nextConfig, nil
}
