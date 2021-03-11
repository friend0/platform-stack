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
	// adds yaml if none present in base filename, handles secret files wrt filepath Ext
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
