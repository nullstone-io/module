package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFiles(t *testing.T) {
	tests := []struct {
		name            string
		outputFile      string
		expectedEnvVars interface{}
	}{
		{
			name:            "basic",
			outputFile:      "05/outputs.tf",
			expectedEnvVars: map[string]EnvVariable{
				"REDSHIFT_USER": { Sensitive: false },
				"REDSHIFT_DB": { Sensitive: false },
				"REDSHIFT_PASSWORD": { Sensitive: true },
				"REDSHIFT_URL": { Sensitive: true },
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tfconfig, err := ParseFiles([]string{filepath.Join("test-fixtures", test.outputFile)})
			require.NoError(t, err, "reading test fixture file")
			manifest := tfconfig.ToManifest()
			assert.Equal(t, test.expectedEnvVars, manifest.EnvVariables)
		})
	}
}

