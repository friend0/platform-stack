package v0beta1

import (
	"github.com/altiscope/platform-stack/pkg/schema/util"
)

const Version string = "stack/v0beta1"

func NewStackConfig() util.VersionedConfig {
	return new(StackConfig)
}

func (config *StackConfig) GetVersion() string {
	return Version
}

type StackConfig struct {
	// Note: Version was not present at the time of stack/v0beta1 release.
	// It is included in this configuration to not break strict unmarshalling in cases where it is specified
	ApiVersion   string                   `yaml:"apiVersion"`
	Components   []ComponentDescription   `yaml:"components"`
	Environments []EnvironmentDescription `yaml:"environments"`
	Stack        StackDescription         `yaml:"stack"`
}

type StackDescription struct {
	Name string `yaml:"name"`
}

type ActivationDescription struct {
	ConfirmWithUser bool   `yaml:"confirm_with_user"`
	Env             string `yaml:"env"`
	Context         string `yaml:"context"`
}

type EnvironmentDescription struct {
	Name       string                `yaml:"name"`
	Activation ActivationDescription `yaml:"activation"`
}

type ComponentDescription struct {
	Name              string                 `yaml:"name"`
	RequiredVariables []string               `yaml:"requiredVariables"`
	Exposable         bool                   `yaml:"exposable"`
	Containers        []ContainerDescription `yaml:"containers"`
	Manifests         []string               `yaml:"manifests"`
}

type ContainerDescription struct {
	Dockerfile string `yaml:"dockerfile"`
	Context    string `yaml:"context"`
	Image      string `yaml:"image"`
}

type ManifestDescription struct {
	Dockerfile string `yaml:"dockerfile"`
	Context    string `yaml:"context"`
	Image      string `yaml:"image"`
}
