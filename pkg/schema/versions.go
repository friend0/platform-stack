/*
Copyright 2019 The Skaffold Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package schema

import (
	"fmt"
	"github.com/altiscope/platform-stack/pkg/schema/latest"
	"github.com/altiscope/platform-stack/pkg/schema/util"
	stackUtils "github.com/altiscope/platform-stack/pkg/schema/util"
	"github.com/altiscope/platform-stack/pkg/schema/v0beta1"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

type APIVersion struct {
	Version string `yaml:"apiVersion" json:"apiVersion"`
}

var VersionList = Versions{
	{v0beta1.Version, v0beta1.NewStackConfig},
	{latest.Version, latest.NewStackConfig},
}

type Version struct {
	APIVersion string
	Factory    func() util.VersionedConfig
}

type Versions []Version

func (v *Versions) Find(apiVersion string) (func() util.VersionedConfig, bool) {
	for _, version := range *v {
		if version.APIVersion == apiVersion {
			return version.Factory, true
		}
	}

	return nil, false
}

var re = regexp.MustCompile(`^stack/v(\d)(?:(alpha|beta)([1-9]?[0-9]))?$`)

// GetSemver parses a string into a Version.
func GetSemver(v string) (semver.Version, error) {
	res := re.FindStringSubmatch(v)
	if res == nil {
		return semver.Version{}, fmt.Errorf("%s is an invalid api version", v)
	}
	if res[2] == "" || res[3] == "" {
		return semver.Parse(fmt.Sprintf("%s.0.0", res[1]))
	}
	return semver.Parse(fmt.Sprintf("%s.0.0-%s.%s", res[1], res[2], res[3]))
}

func ParseConfig(filename string, upgrade bool) (util.VersionedConfig, error) {
	buf, err := stackUtils.ReadStackConfiguration(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read stack config")
	}

	apiVersion := &APIVersion{}
	if err := yaml.Unmarshal(buf, apiVersion); err != nil {
		return nil, fmt.Errorf("parsing api version: %w", err)
	}

	if apiVersion.Version == "" {
		// todo: want to warn, but can break dryrun used by tilt
		//fmt.Printf("Stack configuration missing version - treating config as `stack/v0beta1`\n")
		apiVersion.Version = "stack/v0beta1"
	}

	factory, present := VersionList.Find(apiVersion.Version)
	if !present {
		return nil, fmt.Errorf("unknown api version: '%s'", apiVersion.Version)
	}

	// Remove all top-level keys starting with `.` so they can be used as YAML anchors
	parsed := make(map[string]interface{})
	if err := yaml.UnmarshalStrict(buf, parsed); err != nil {
		return nil, fmt.Errorf("unable to parse YAML: %w", err)
	}
	for field := range parsed {
		if strings.HasPrefix(field, ".") {
			delete(parsed, field)
		}
	}
	buf, err = yaml.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("unable to re-marshal YAML without dotted keys: %w", err)
	}

	cfg := factory()
	if err := yaml.UnmarshalStrict(buf, cfg); err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	if upgrade && cfg.GetVersion() != latest.Version {
		cfg, err = upgradeToLatest(cfg)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// upgradeToLatest upgrades a configuration to the latest version.
func upgradeToLatest(vc util.VersionedConfig) (util.VersionedConfig, error) {
	var err error

	// first, check to make sure config version isn't too new
	version, err := GetSemver(vc.GetVersion())
	if err != nil {
		return nil, fmt.Errorf("parsing api version: %w", err)
	}

	latestSemanticVersion, err := GetSemver(latest.Version)
	if err != nil {
		return nil, err
	}

	if version.EQ(latestSemanticVersion) {
		return vc, nil
	}
	if version.GT(latestSemanticVersion) {
		return nil, fmt.Errorf("config version %s is too new for this version: upgrade stack CLI", vc.GetVersion())
	}

	for vc.GetVersion() != latest.Version {
		vc, err = vc.Upgrade()
		if err != nil {
			return nil, fmt.Errorf("transforming stack config: %v", err)
		}
	}

	return vc, nil
}
