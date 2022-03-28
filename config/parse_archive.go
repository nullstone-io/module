package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/nullstone-io/module/scan"
)

var (
	validTfFileExts = map[string]bool{
		".tf":      true,
		".tf.json": true,
	}
)

func ParseArchive(archiveData []byte, ext string) (*InternalTfConfig, error) {
	scanner, ok := scan.ArchiveScanners[strings.TrimPrefix(ext, ".")]
	if !ok {
		return nil, fmt.Errorf("unsupported archive %q", ext)
	}

	root := &InternalTfConfig{}
	err := scanner.Scan(bytes.NewBuffer(archiveData), func(fullname string, r io.Reader) error {
		dir, filename := filepath.Split(fullname)
		if dir != "" {
			// We skip nested files when parsing manifest
			return nil
		}
		// remove the base path, just check for file name
		if strings.ToLower(filepath.Base(filename)) == "readme.md" {
			raw, err := ioutil.ReadAll(r)
			if err != nil {
				return fmt.Errorf("error reading the README.md file: %w", err)
			}
			root.Readme = string(raw)
			return nil
		}
		ext := filepath.Ext(filename)
		if _, ok := validTfFileExts[ext]; !ok {
			// Not a TF file, we can ignore from parsing
			return nil
		}
		raw, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("error reading archive file %q: %w", filename, err)
		}
		return ParseFileContents(root, raw, filename)
	})
	return root, err
}
