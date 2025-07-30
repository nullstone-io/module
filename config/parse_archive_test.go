package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArchive(t *testing.T) {
	tests := []struct {
		name         string
		archiveFile  string
		expectedFile string
	}{
		{
			name:         "tgz-basic",
			archiveFile:  "01/module.tgz",
			expectedFile: "01/expected.json",
		},
		{
			name:         "tgz-with-subdirs",
			archiveFile:  "02/module.tgz",
			expectedFile: "02/expected.json",
		},
		{
			name:         "tgz-network",
			archiveFile:  "03/module.tgz",
			expectedFile: "03/expected.json",
		},
		{
			name:         "zip-network",
			archiveFile:  "03/module.zip",
			expectedFile: "03/expected.json",
		},
		{
			name:         "tgz-with-a-nontf-file",
			archiveFile:  "04/module.tgz",
			expectedFile: "04/expected.json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("test-fixtures", test.archiveFile))
			require.NoError(t, err, "reading test fixture archive")
			cfg, err := ParseArchive(raw, ArchiveExt(test.archiveFile))
			require.NoError(t, err)
			wantRaw, err := os.ReadFile(filepath.Join("test-fixtures", test.expectedFile))
			require.NoError(t, err)
			var want Manifest
			require.NoError(t, json.Unmarshal(wantRaw, &want))

			got := cfg.ToManifest()
			assert.Equal(t, want, got)
		})
	}
}

func TestParseArchive_Meta(t *testing.T) {
	tests := []struct {
		name               string
		archiveFile        string
		expectedReadmeFile string
		expectedChangelog  string
	}{
		{
			name:               "fake-module-with-readme",
			archiveFile:        "07/module.tgz",
			expectedReadmeFile: "07/README.md",
			expectedChangelog:  "07/CHANGELOG.md",
		},
	}

	mustReadFileAsString := func(t *testing.T, fullname string) string {
		wantChangelogRaw, err := os.ReadFile(fullname)
		require.NoError(t, err)
		return string(wantChangelogRaw)
	}

	for _, test := range tests {
		raw, err := os.ReadFile(filepath.Join("test-fixtures", test.archiveFile))
		require.NoError(t, err, "reading test fixture archive")
		wantReadme := mustReadFileAsString(t, filepath.Join("test-fixtures", test.expectedReadmeFile))
		wantChangelog := mustReadFileAsString(t, filepath.Join("test-fixtures", test.expectedChangelog))

		cfg, err := ParseArchive(raw, ArchiveExt(test.archiveFile))
		require.NoError(t, err)
		assert.Equal(t, wantReadme, cfg.Readme)
		assert.Equal(t, wantChangelog, cfg.Changelog)
	}
}
