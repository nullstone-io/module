package platformdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// KindEnv is the platform data kind carrying the resolved environment of an application workspace.
const KindEnv = "env"

// envVariableKeyRegex mirrors the key rule enforced by terraform-provider-ns for env variables.
var envVariableKeyRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// EnvV1 is version 1 of the `env` kind: one entry per env variable name.
//
// Every variable is exactly one of:
//   - a resolved value:   {template?, value}
//   - a sensitive value:  {template?, sensitive: true}            (the value is never carried)
//   - a reference:        {template?, ref: {type, ...}}           (resolved by the platform at runtime)
//
// Each variable may also carry its source layer (standard, cloud, otel, capability, user).
// A reference of type "secret" is sensitive by definition; k8s refs are not.
// Invariant: no secret value is ever carried. The `ns_env_variables` data source promotes any
// variable that interpolates a secret into `secrets`, and this schema only marks those as sensitive.
type EnvV1 struct {
	// Platform names the runtime (see LookupPlatform). Optional; empty for records produced by the legacy adapter.
	Platform  string                   `json:"platform,omitempty"`
	Variables map[string]EnvV1Variable `json:"variables"`
}

type EnvV1Variable struct {
	// Template is the raw value before interpolation (e.g. "{{ NULLSTONE_ENV }}-db"). Optional.
	Template string `json:"template,omitempty"`
	// Value is the resolved value. Must be absent when Sensitive or Ref is set.
	Value string `json:"value,omitempty"`
	// Sensitive marks a secret whose value is not carried.
	Sensitive bool `json:"sensitive,omitempty"`
	// Ref describes where the value is resolved from at runtime.
	Ref *EnvV1Ref `json:"ref,omitempty"`
	// Source is the layer that supplied this variable (see Source* constants). Optional.
	Source string `json:"source,omitempty"`
	// Capability identifies the capability when Source is "capability". Optional.
	Capability string `json:"capability,omitempty"`
}

// Source layers, lowest to highest precedence.
const (
	SourceStandard   = "standard"   // NULLSTONE_* built-ins
	SourceCloud      = "cloud"      // cloud platform built-ins (GOOGLE_*, AWS_*, AZURE_*)
	SourceOtel       = "otel"       // OTEL_* wiring
	SourceCapability = "capability" // env/secrets emitted by a capability module
	SourceUser       = "user"       // var.env_vars / var.secrets
)

// Sources returns the layers in precedence order (lowest first).
func Sources() []string {
	return []string{SourceStandard, SourceCloud, SourceOtel, SourceCapability, SourceUser}
}

func knownSource(s string) bool {
	for _, k := range Sources() {
		if k == s {
			return true
		}
	}
	return false
}

// Ref types
const (
	RefTypeSecret           = "secret"             // cloud secret store reference (ARN, GCP secret id, ...)
	RefTypeK8sField         = "k8s_field"          // k8s fieldRef
	RefTypeK8sConfigMap     = "k8s_config_map"     // k8s configMapKeyRef
	RefTypeK8sResourceField = "k8s_resource_field" // k8s resourceFieldRef
	RefTypeK8sFileKey       = "k8s_file_key"       // k8s fileKeyRef
)

// EnvV1Ref is a tagged union keyed by Type; only the fields for that type are set.
type EnvV1Ref struct {
	Type string `json:"type"`

	// secret
	Id string `json:"id,omitempty"`

	// k8s_field
	ApiVersion string `json:"api_version,omitempty"`
	FieldPath  string `json:"field_path,omitempty"`

	// k8s_config_map
	Name     string `json:"name,omitempty"`
	Key      string `json:"key,omitempty"` // also k8s_file_key
	Optional bool   `json:"optional,omitempty"`

	// k8s_resource_field
	Resource  string `json:"resource,omitempty"`
	Container string `json:"container,omitempty"`
	Divisor   string `json:"divisor,omitempty"`

	// k8s_file_key
	VolumeName string `json:"volume_name,omitempty"`
	Path       string `json:"path,omitempty"`
}

// IsSensitive reports whether the variable's value must never be shown.
func (v EnvV1Variable) IsSensitive() bool {
	return v.Sensitive || (v.Ref != nil && v.Ref.Type == RefTypeSecret)
}

// ParseEnvV1 decodes and validates a raw payload.
func ParseEnvV1(raw json.RawMessage) (EnvV1, error) {
	var env EnvV1
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&env); err != nil {
		return EnvV1{}, fmt.Errorf("unable to decode: %w", err)
	}
	if err := env.Validate(); err != nil {
		return EnvV1{}, err
	}
	return env, nil
}

// Validate enforces the invariants of the env kind.
func (e EnvV1) Validate() error {
	if e.Variables == nil {
		return errors.New("variables is required (use an empty object when there are none)")
	}
	var errs []error
	if e.Platform != "" {
		if _, ok := LookupPlatform(e.Platform); !ok {
			errs = append(errs, fmt.Errorf("unknown platform %q", e.Platform))
		}
	}
	for _, key := range e.sortedKeys() {
		if !envVariableKeyRegex.MatchString(key) {
			errs = append(errs, fmt.Errorf("%q is not a valid env variable name", key))
		}
		for _, err := range e.Variables[key].validate() {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
		}
	}
	return errors.Join(errs...)
}

func (v EnvV1Variable) validate() []error {
	var errs []error
	if v.Sensitive && v.Value != "" {
		errs = append(errs, errors.New("a sensitive variable cannot carry a value"))
	}
	if v.Source != "" && !knownSource(v.Source) {
		errs = append(errs, fmt.Errorf("unknown source %q", v.Source))
	}
	if v.Capability != "" && v.Source != SourceCapability {
		errs = append(errs, errors.New("capability is only valid when source is \"capability\""))
	}
	if v.Ref == nil {
		return errs
	}
	if v.Value != "" {
		errs = append(errs, errors.New("a variable cannot have both a value and a ref"))
	}
	r := v.Ref
	require := func(field, val string) {
		if strings.TrimSpace(val) == "" {
			errs = append(errs, fmt.Errorf("ref type %q requires %s", r.Type, field))
		}
	}
	switch r.Type {
	case RefTypeSecret:
		require("id", r.Id)
	case RefTypeK8sField:
		require("field_path", r.FieldPath)
	case RefTypeK8sConfigMap:
		require("name", r.Name)
		require("key", r.Key)
	case RefTypeK8sResourceField:
		require("resource", r.Resource)
	case RefTypeK8sFileKey:
		require("volume_name", r.VolumeName)
		require("path", r.Path)
		require("key", r.Key)
	case "":
		errs = append(errs, errors.New("ref requires a type"))
	default:
		errs = append(errs, fmt.Errorf("unknown ref type %q", r.Type))
	}
	return errs
}

func (e EnvV1) sortedKeys() []string {
	keys := make([]string, 0, len(e.Variables))
	for k := range e.Variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Keys returns every env var name in this record, sorted.
func (e EnvV1) Keys() []string {
	return e.sortedKeys()
}

// EnvV1Schema implements Schema for (env, 1).
type EnvV1Schema struct{}

func (EnvV1Schema) Validate(raw json.RawMessage) error {
	_, err := ParseEnvV1(raw)
	return err
}

// Merge requires exactly one instance: an application workspace has one environment.
func (EnvV1Schema) Merge(instances []json.RawMessage) (json.RawMessage, error) {
	switch len(instances) {
	case 0:
		return nil, errors.New("no env platform data found")
	case 1:
		if _, err := ParseEnvV1(instances[0]); err != nil {
			return nil, err
		}
		return instances[0], nil
	default:
		return nil, fmt.Errorf("expected exactly one env platform data record, found %d", len(instances))
	}
}
