package latest

import (
	"errors"
	"github.com/altiscope/platform-stack/pkg/schema/util"
)

// Upgrade upgrades a configuration to the next version.
func (c *StackConfig) Upgrade() (util.VersionedConfig, error) {
	return nil, errors.New("not implemented yet")
}
