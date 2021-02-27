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

package util

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

type VersionedConfig interface {
	GetVersion() string
	Upgrade() (VersionedConfig, error)
}

func ReadStackConfiguration(filename string) ([]byte, error) {
	workingFilePath := filename
	// adds yaml if none present in base filename, handles secret files wrt  filepath Ext
	if filepath.Ext(strings.TrimPrefix(filepath.Base(filename), ".")) == "" {
		workingFilePath += ".yaml"
	}
	contents, err := ioutil.ReadFile(workingFilePath)
	if err != nil {
		if filepath.Ext(filename) == "yaml" {
			contents, errIgnored := ioutil.ReadFile(filename + ".yml")
			if errIgnored != nil {
				return nil, err
			}
			return contents, nil
		}
	}
	return contents, err
}
