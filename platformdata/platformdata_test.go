package platformdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixture(t *testing.T, name string) json.RawMessage {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("fixtures", name))
	require.NoError(t, err)
	return raw
}

func TestRegistry(t *testing.T) {
	assert.Equal(t, []string{"env"}, Kinds())
	assert.True(t, Known("env"))
	assert.False(t, Known("metrics"))
	assert.Equal(t, 1, LatestVersion("env"))
	assert.Equal(t, 0, LatestVersion("metrics"))

	s, ok := Lookup("env", 1)
	assert.True(t, ok)
	assert.IsType(t, EnvV1Schema{}, s)
	_, ok = Lookup("env", 2)
	assert.False(t, ok)
	_, ok = Lookup("metrics", 1)
	assert.False(t, ok)
}

func TestValidateEnvelope(t *testing.T) {
	tests := []struct {
		name         string
		envelope     Envelope
		wantWarnings int
		wantErr      string
	}{
		{
			name:     "known kind and version, valid data",
			envelope: Envelope{Kind: "env", Version: 1, Data: fixture(t, "env_v1_valid.json")},
		},
		{
			name:     "known kind and version, invalid data",
			envelope: Envelope{Kind: "env", Version: 1, Data: fixture(t, "env_v1_secret_overlap.json")},
			wantErr:  `"DATABASE_PASSWORD" cannot be both a variable and a secret`,
		},
		{
			name:         "unknown kind warns",
			envelope:     Envelope{Kind: "metrics", Version: 1, Data: json.RawMessage(`{"anything":true}`)},
			wantWarnings: 1,
		},
		{
			name:         "unknown version warns",
			envelope:     Envelope{Kind: "env", Version: 99, Data: json.RawMessage(`{}`)},
			wantWarnings: 1,
		},
		{
			name:     "missing kind",
			envelope: Envelope{Version: 1, Data: json.RawMessage(`{}`)},
			wantErr:  "kind is required",
		},
		{
			name:     "bad version",
			envelope: Envelope{Kind: "env", Version: 0, Data: json.RawMessage(`{}`)},
			wantErr:  "version must be a positive integer",
		},
		{
			name:     "data not an object",
			envelope: Envelope{Kind: "env", Version: 1, Data: json.RawMessage(`[1,2]`)},
			wantErr:  "data must be a JSON object",
		},
		{
			name:     "data null",
			envelope: Envelope{Kind: "env", Version: 1, Data: json.RawMessage(`null`)},
			wantErr:  "data must be a JSON object",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			warnings, err := ValidateEnvelope(test.envelope)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, warnings, test.wantWarnings)
		})
	}
}

func TestParseEnvV1(t *testing.T) {
	t.Run("valid fixture", func(t *testing.T) {
		env, err := ParseEnvV1(fixture(t, "env_v1_valid.json"))
		require.NoError(t, err)
		assert.Equal(t, "prod-db", env.Variables["DATABASE_NAME"].Value)
		assert.Equal(t, "{{ NULLSTONE_ENV }}-db", env.Variables["DATABASE_NAME"].Template)
		assert.Equal(t, "", env.Variables["LOG_LEVEL"].Template)
		assert.ElementsMatch(t, []string{"DATABASE_PASSWORD", "API_TOKEN"}, env.SecretKeys)
		require.NotNil(t, env.K8s)
		assert.Equal(t, "status.podIP", env.K8s.FieldRefs["POD_IP"].FieldPath)
		assert.Equal(t, []string{
			"API_TOKEN", "CERT", "CPU_LIMIT", "DATABASE_NAME", "DATABASE_PASSWORD",
			"FEATURE_FLAGS", "LOG_LEVEL", "NULLSTONE_ENV", "POD_IP",
		}, env.Keys())
	})

	t.Run("bad keys reports every problem", func(t *testing.T) {
		_, err := ParseEnvV1(fixture(t, "env_v1_bad_keys.json"))
		require.Error(t, err)
		msg := err.Error()
		assert.Contains(t, msg, `variables: "9LIVES" is not a valid env variable name`)
		assert.Contains(t, msg, `variables: "HAS-DASH" is not a valid env variable name`)
		assert.Contains(t, msg, `secret_keys: "OK_KEY" is listed more than once`)
		assert.Contains(t, msg, `secret_refs: "NOT_A_SECRET" must also be listed in secret_keys`)
		assert.Contains(t, msg, `secret_refs: "OK_KEY" has an empty reference`)
	})

	t.Run("variables required", func(t *testing.T) {
		_, err := ParseEnvV1(json.RawMessage(`{"secret_keys":[]}`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "variables is required")
	})

	t.Run("empty variables ok", func(t *testing.T) {
		env, err := ParseEnvV1(json.RawMessage(`{"variables":{}}`))
		require.NoError(t, err)
		assert.Empty(t, env.Keys())
	})

	t.Run("unknown fields rejected", func(t *testing.T) {
		_, err := ParseEnvV1(json.RawMessage(`{"variables":{},"secrets":{"X":"leak"}}`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown field "secrets"`)
	})

	t.Run("k8s ref conflicts with variable", func(t *testing.T) {
		_, err := ParseEnvV1(json.RawMessage(`{"variables":{"POD_IP":{"value":"x"}},"k8s":{"field_refs":{"POD_IP":{"field_path":"status.podIP"}}}}`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), `k8s.field_refs: "POD_IP" cannot be both a variable and a runtime reference`)
	})
}

func TestEnvV1Schema_Merge(t *testing.T) {
	valid := fixture(t, "env_v1_valid.json")
	schema := EnvV1Schema{}

	merged, err := schema.Merge([]json.RawMessage{valid})
	require.NoError(t, err)
	assert.JSONEq(t, string(valid), string(merged))

	_, err = schema.Merge(nil)
	assert.EqualError(t, err, "no env platform data found")

	_, err = schema.Merge([]json.RawMessage{valid, valid})
	assert.EqualError(t, err, "expected exactly one env platform data record, found 2")

	_, err = schema.Merge([]json.RawMessage{fixture(t, "env_v1_secret_overlap.json")})
	require.Error(t, err)
}
