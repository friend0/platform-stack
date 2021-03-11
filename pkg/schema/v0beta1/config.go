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
	ApiVersion   string                   `yaml:"apiVersion" json:"apiVersion"`
	Components   []ComponentDescription   `yaml:"components" json:"components"`
	Environments []EnvironmentDescription `yaml:"environments" json:"environments"`
	Stack        StackDescription         `yaml:"stack" json:"stack"`
}

type StackDescription struct {
	Name string `yaml:"name" json:"name"`
}

type ActivationDescription struct {
	ConfirmWithUser bool   `yaml:"confirmWithUser" json:"confirmWithUser"`
	Env             string `yaml:"env" json:"env"`
	Context         string `yaml:"context" json:"context"`
}

type EnvironmentDescription struct {
	Name       string                `yaml:"name" json:"name"`
	Activation ActivationDescription `yaml:"activation" json:"activation"`
}

type ComponentDescription struct {
	Name              string                 `yaml:"name" json:"name"`
	RequiredVariables []string               `yaml:"requiredVariables" json:"requiredVariables"`
	Exposable         bool                   `yaml:"exposable" json:"exposable"`
	Containers        []ContainerDescription `yaml:"containers" json:"containers"`
	Manifests         []string               `yaml:"manifests" json:"manifests"`
	TemplateConfig    []string               `yaml:"templateConfig" json:"templateConfig"`
}

type ContainerDescription struct {
	Dockerfile   string   `yaml:"dockerfile" json:"dockerfile"`
	Context      string   `yaml:"context" json:"context"`
	Image        string   `yaml:"image" json:"image"`
	Environments []string `yaml:"environments,omitempty" json:"environments,omitempty"`
}

type ManifestDescription struct {
	Dockerfile string `yaml:"dockerfile" json:"dockerfile"`
	Context    string `yaml:"context" json:"context"`
	Image      string `yaml:"image" json:"image"`
}

type Config struct {
	Components   []ComponentDescription   `yaml:"components" json:"components"`
	Environments []EnvironmentDescription `yaml:"environments" json:"environments"`
	Stack        StackDescription         `yaml:"stack" json:"stack"`
}
