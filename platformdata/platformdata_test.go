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
			wantErr:  "DATABASE_PASSWORD: a sensitive variable cannot carry a value",
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

		assert.True(t, env.Variables["DATABASE_PASSWORD"].IsSensitive())
		assert.Empty(t, env.Variables["DATABASE_PASSWORD"].Value)
		assert.True(t, env.Variables["API_TOKEN"].IsSensitive())
		assert.True(t, env.Variables["GCP_TOKEN"].IsSensitive(), "secret refs are sensitive even without the flag")
		assert.False(t, env.Variables["POD_IP"].IsSensitive())
		assert.False(t, env.Variables["EMPTY"].IsSensitive())

		assert.Equal(t, &EnvV1Ref{Type: RefTypeK8sField, ApiVersion: "v1", FieldPath: "status.podIP"}, env.Variables["POD_IP"].Ref)
		assert.Equal(t, &EnvV1Ref{Type: RefTypeK8sConfigMap, Name: "flags", Key: "flags.json", Optional: true}, env.Variables["FEATURE_FLAGS"].Ref)
		assert.Equal(t, &EnvV1Ref{Type: RefTypeK8sResourceField, Resource: "limits.cpu", Container: "app"}, env.Variables["CPU_LIMIT"].Ref)
		assert.Equal(t, &EnvV1Ref{Type: RefTypeK8sFileKey, VolumeName: "certs", Path: "/etc/certs", Key: "tls.crt"}, env.Variables["CERT"].Ref)
		assert.Equal(t, []string{
			"API_TOKEN", "CERT", "CPU_LIMIT", "DATABASE_NAME", "DATABASE_PASSWORD", "EMPTY",
			"FEATURE_FLAGS", "GCP_TOKEN", "LOG_LEVEL", "NULLSTONE_ENV", "POD_IP",
		}, env.Keys())
	})

	t.Run("bad keys reports every problem", func(t *testing.T) {
		_, err := ParseEnvV1(fixture(t, "env_v1_bad_keys.json"))
		require.Error(t, err)
		msg := err.Error()
		assert.Contains(t, msg, `"9LIVES" is not a valid env variable name`)
		assert.Contains(t, msg, `"HAS-DASH" is not a valid env variable name`)
		assert.Contains(t, msg, `BOTH: a variable cannot have both a value and a ref`)
		assert.Contains(t, msg, `EMPTY_REF_ID: ref type "secret" requires id`)
		assert.Contains(t, msg, `NO_REF_TYPE: ref requires a type`)
		assert.Contains(t, msg, `BAD_REF_TYPE: unknown ref type "vault"`)
		assert.Contains(t, msg, `CM_MISSING: ref type "k8s_config_map" requires key`)
		assert.Contains(t, msg, `FILE_MISSING: ref type "k8s_file_key" requires volume_name`)
		assert.Contains(t, msg, `FILE_MISSING: ref type "k8s_file_key" requires path`)
	})

	t.Run("variables required", func(t *testing.T) {
		_, err := ParseEnvV1(json.RawMessage(`{}`))
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

	t.Run("sensitive with ref is fine", func(t *testing.T) {
		env, err := ParseEnvV1(json.RawMessage(`{"variables":{"T":{"sensitive":true,"ref":{"type":"secret","id":"x"}}}}`))
		require.NoError(t, err)
		assert.True(t, env.Variables["T"].IsSensitive())
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
