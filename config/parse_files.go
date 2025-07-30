package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func ParseFiles(files []string) (*InternalTfConfig, error) {
	root := &InternalTfConfig{}
	for _, file := range files {
		filename, raw, err := ReadFile(file)
		if err != nil {
			return nil, err
		}
		if err := root.ReadInFile(filename, raw); err != nil {
			return nil, err
		}
	}
	return root, nil
}

func ReadFile(fullname string) (string, []byte, error) {
	filename := filepath.Base(fullname)
	raw, err := os.ReadFile(fullname)
	if err != nil {
		return filename, nil, fmt.Errorf("error reading %q: %s", fullname, err)
	}
	return filename, raw, nil
}
