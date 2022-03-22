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
		expectedOutputs map[string]Output
		expectedEnvVars map[string]EnvVariable
	}{
		{
			name:            "parse env and secrets",
			outputFile:      "05/outputs.tf",
			expectedOutputs: map[string]Output{},
			expectedEnvVars: map[string]EnvVariable{
				"REDSHIFT_USER":     {Sensitive: false},
				"REDSHIFT_DB":       {Sensitive: false},
				"REDSHIFT_PASSWORD": {Sensitive: true},
				"REDSHIFT_URL":      {Sensitive: true},
			},
		},
		{
			name:       "fall back to regular outputs",
			outputFile: "06/outputs.tf",
			expectedOutputs: map[string]Output{
				"db_instance_arn":       {Type: "string", Description: "ARN of the Postgres instance", Sensitive: false},
				"db_master_secret_name": {Type: "string", Description: "The name of the secret in AWS Secrets Manager containing the password", Sensitive: false},
			},
			expectedEnvVars: map[string]EnvVariable{},
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
