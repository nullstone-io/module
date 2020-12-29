package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tmccombs/hcl2json/convert"
)

func ParseDir(dir string) (*InternalTfConfig, error) {
	files, err := ReadDir(dir)
	if err != nil {
		return nil, err
	}
	return ParseFiles(files)
}

func ParseFiles(files []string) (*InternalTfConfig, error) {
	root := &InternalTfConfig{}
	for _, file := range files {
		raw, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to read %q: %s", file, err)
		}
		rawJson, err := convert.Bytes(raw, filepath.Base(file), convert.Options{})
		if err != nil {
			return nil, fmt.Errorf("unable to convert hcl to json %q: %s", file, err)
		}
		var curManifest InternalTfConfig
		if err := json.Unmarshal(rawJson, &curManifest); err != nil {
			return nil, fmt.Errorf("unable to unmarshal raw converted json %q: %s", file, err)
		}
		root.MergeIn(curManifest)
	}
	return root, nil
}
