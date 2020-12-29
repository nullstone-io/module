package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternal(t *testing.T) {
	tests := []string{
		"01",
		"02",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			files, err := ReadDir(filepath.Join("test-fixtures", test))
			require.NoError(t, err)
			cfg, err := Parse(files)
			require.NoError(t, err)
			wantRaw, err := ioutil.ReadFile(filepath.Join("test-fixtures", test, "expected.json"))
			require.NoError(t, err)
			var want Manifest
			require.NoError(t, json.Unmarshal(wantRaw, &want))

			got := cfg.ToManifest()
			assert.Equal(t, want, got)
		})
	}
}
