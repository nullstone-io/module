package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tmccombs/hcl2json/convert"
)

func ParseFiles(files []string) (*InternalTfConfig, error) {
	root := &InternalTfConfig{}
	for _, file := range files {
		raw, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to read %q: %s", file, err)
		}
		if err := ParseFileContents(root, raw, filepath.Base(file)); err != nil {
			return nil, err
		}
	}
	return root, nil
}

func ParseFileContents(root *InternalTfConfig, raw []byte, filename string) error {
	rawJson, err := convert.Bytes(raw, filename, convert.Options{})
	if err != nil {
		return fmt.Errorf("unable to convert hcl to json %q: %s", filename, err)
	}
	var curManifest InternalTfConfig
	if err := json.Unmarshal(rawJson, &curManifest); err != nil {
		return fmt.Errorf("unable to unmarshal raw converted json %q: %s", filename, err)
	}
	root.MergeIn(curManifest)
	return nil
}
