package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDir(t *testing.T) {
	tests := []struct {
		name              string
		readmeContents    string
		changelogContents string
	}{
		{
			name:           "01",
			readmeContents: "# Readme Title",
			changelogContents: `# 0.1.0 (Jul 30, 2025)
* Initial draft
`,
		},
		{
			name:              "02",
			readmeContents:    "",
			changelogContents: ``,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, err := ParseDir(filepath.Join("test-fixtures", test.name))
			require.NoError(t, err)
			wantRaw, err := os.ReadFile(filepath.Join("test-fixtures", test.name, "expected.json"))
			require.NoError(t, err)
			var want Manifest
			require.NoError(t, json.Unmarshal(wantRaw, &want))

			assert.Equal(t, test.readmeContents, cfg.Readme)
			assert.Equal(t, test.changelogContents, cfg.Changelog)
			got := cfg.ToManifest()
			assert.Equal(t, want, got)
		})
	}
}
