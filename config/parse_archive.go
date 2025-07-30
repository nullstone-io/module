package config

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/nullstone-io/module/scan"
)

func ParseArchive(archiveData []byte, ext string) (*InternalTfConfig, error) {
	scanner, ok := scan.ArchiveScanners[strings.TrimPrefix(ext, ".")]
	if !ok {
		return nil, fmt.Errorf("unsupported archive %q", ext)
	}

	root := &InternalTfConfig{}
	err := scanner.Scan(bytes.NewBuffer(archiveData), func(fullname string, r io.Reader) error {
		filename, raw, err := ReadArchiveFile(fullname, r)
		if err != nil {
			return err
		}
		return root.ReadInFile(filename, raw)
	})
	return root, err
}

func ReadArchiveFile(fullname string, r io.Reader) (string, []byte, error) {
	dir, filename := filepath.Split(fullname)
	dir = strings.TrimPrefix(dir, "./")
	if dir != "" {
		// We skip nested files when parsing manifest
		return filename, nil, nil
	}
	raw, err := io.ReadAll(r)
	if err != nil {
		return filename, nil, fmt.Errorf("error reading %q: %s", fullname, err)
	}
	return filename, raw, nil
}
