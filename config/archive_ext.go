package config

import (
	"path/filepath"
	"strings"
)

func ArchiveExt(filename string) string {
	ext := filepath.Ext(filename)
	// filepath.Ext doesn't include .tar in .tar.gz, add it
	if strings.HasSuffix(strings.TrimSuffix(filename, ext), ".tar") {
		ext = ".tar" + ext
	}
	ext = strings.TrimPrefix(ext, ".")
	return ext
}
