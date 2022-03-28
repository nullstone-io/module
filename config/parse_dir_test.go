package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDir(t *testing.T) {
	tests := []struct {
		name           string
		readmeContents string
	}{
		{
			name:           "01",
			readmeContents: "# Readme Title",
		},
		{
			name:           "02",
			readmeContents: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, err := ParseDir(filepath.Join("test-fixtures", test.name))
			require.NoError(t, err)
			wantRaw, err := ioutil.ReadFile(filepath.Join("test-fixtures", test.name, "expected.json"))
			require.NoError(t, err)
			var want Manifest
			require.NoError(t, json.Unmarshal(wantRaw, &want))

			assert.Equal(t, test.readmeContents, cfg.Readme)
			got := cfg.ToManifest()
			assert.Equal(t, want, got)
		})
	}
}
