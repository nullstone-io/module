package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

func ReadDir(dir string) ([]string, error) {
	fs := afero.Afero{Fs: afero.NewOsFs()}
	infos, err := fs.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("module directory %s does not exist or cannot be read", dir)
	}

	files := make([]string, 0)
	for _, info := range infos {
		if info.IsDir() {
			// We only care about files
			continue
		}

		name := info.Name()
		ext := fileExt(name)
		if ext == "" || isIgnoredFile(name) {
			continue
		}

		files = append(files, filepath.Join(dir, name))
	}
	return files, nil
}

// fileExt returns the Terraform configuration extension of the given
// path, or a blank string if it is not a recognized extension.
func fileExt(path string) string {
	if strings.HasSuffix(path, ".tf") {
		return ".tf"
	} else if strings.HasSuffix(path, ".tf.json") {
		return ".tf.json"
	} else {
		return ""
	}
}

// IsIgnoredFile returns true if the given filename (which must not have a
// directory path ahead of it) should be ignored as e.g. an editor swap file.
func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, ".") || // Unix-like hidden files
		strings.HasSuffix(name, "~") || // vim
		strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#") // emacs
}
