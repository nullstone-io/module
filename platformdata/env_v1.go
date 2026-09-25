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

// EnvV1 is version 1 of the `env` kind.
//
// Invariants:
//   - Variables never contain secret values. The `ns_env_variables` data source promotes any
//     variable that interpolates a secret into `secrets`, and this schema only carries secret key names.
//   - SecretKeys and SecretRefs carry names/references only, never values.
type EnvV1 struct {
	// Variables maps env var name to its resolved value and, optionally, the raw template it was resolved from.
	Variables map[string]EnvV1Variable `json:"variables"`
	// SecretKeys lists env var names whose values are secrets. Values are never carried.
	SecretKeys []string `json:"secret_keys,omitempty"`
	// SecretRefs maps env var name to a reference to an existing secret (ARN, GCP secret id, ...).
	SecretRefs map[string]string `json:"secret_refs,omitempty"`
	// K8s carries Kubernetes valueFrom-style references.
	K8s *EnvV1K8sRefs `json:"k8s,omitempty"`
}

type EnvV1Variable struct {
	// Template is the raw value before interpolation (e.g. "{{ NULLSTONE_ENV }}-db"). Optional.
	Template string `json:"template,omitempty"`
	// Value is the resolved value.
	Value string `json:"value"`
}

type EnvV1K8sRefs struct {
	FieldRefs         map[string]EnvV1FieldRef         `json:"field_refs,omitempty"`
	ConfigMapRefs     map[string]EnvV1ConfigMapRef     `json:"config_map_refs,omitempty"`
	ResourceFieldRefs map[string]EnvV1ResourceFieldRef `json:"resource_field_refs,omitempty"`
	FileKeyRefs       map[string]EnvV1FileKeyRef       `json:"file_key_refs,omitempty"`
}

type EnvV1FieldRef struct {
	ApiVersion string `json:"api_version,omitempty"`
	FieldPath  string `json:"field_path"`
}

type EnvV1ConfigMapRef struct {
	Name     string `json:"name"`
	Key      string `json:"key"`
	Optional bool   `json:"optional,omitempty"`
}

type EnvV1ResourceFieldRef struct {
	Resource  string `json:"resource"`
	Container string `json:"container,omitempty"`
	Divisor   string `json:"divisor,omitempty"`
}

type EnvV1FileKeyRef struct {
	VolumeName string `json:"volume_name"`
	Path       string `json:"path"`
	Key        string `json:"key"`
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
	var errs []error
	if e.Variables == nil {
		errs = append(errs, errors.New("variables is required (use an empty object when there are none)"))
	}
	for key := range e.Variables {
		if !envVariableKeyRegex.MatchString(key) {
			errs = append(errs, fmt.Errorf("variables: %q is not a valid env variable name", key))
		}
	}
	secretKeys := map[string]bool{}
	for _, key := range e.SecretKeys {
		if !envVariableKeyRegex.MatchString(key) {
			errs = append(errs, fmt.Errorf("secret_keys: %q is not a valid env variable name", key))
			continue
		}
		if secretKeys[key] {
			errs = append(errs, fmt.Errorf("secret_keys: %q is listed more than once", key))
		}
		secretKeys[key] = true
		if _, ok := e.Variables[key]; ok {
			errs = append(errs, fmt.Errorf("%q cannot be both a variable and a secret", key))
		}
	}
	for key, ref := range e.SecretRefs {
		if !secretKeys[key] {
			errs = append(errs, fmt.Errorf("secret_refs: %q must also be listed in secret_keys", key))
		}
		if strings.TrimSpace(ref) == "" {
			errs = append(errs, fmt.Errorf("secret_refs: %q has an empty reference", key))
		}
	}
	if e.K8s != nil {
		for key := range e.K8s.FieldRefs {
			errs = append(errs, e.checkRefKey("k8s.field_refs", key)...)
		}
		for key := range e.K8s.ConfigMapRefs {
			errs = append(errs, e.checkRefKey("k8s.config_map_refs", key)...)
		}
		for key := range e.K8s.ResourceFieldRefs {
			errs = append(errs, e.checkRefKey("k8s.resource_field_refs", key)...)
		}
		for key := range e.K8s.FileKeyRefs {
			errs = append(errs, e.checkRefKey("k8s.file_key_refs", key)...)
		}
	}
	return errors.Join(errs...)
}

func (e EnvV1) checkRefKey(section, key string) []error {
	var errs []error
	if !envVariableKeyRegex.MatchString(key) {
		errs = append(errs, fmt.Errorf("%s: %q is not a valid env variable name", section, key))
	}
	if _, ok := e.Variables[key]; ok {
		errs = append(errs, fmt.Errorf("%s: %q cannot be both a variable and a runtime reference", section, key))
	}
	return errs
}

// Keys returns every env var name known to this record (variables, secrets, refs), sorted.
func (e EnvV1) Keys() []string {
	set := map[string]bool{}
	for k := range e.Variables {
		set[k] = true
	}
	for _, k := range e.SecretKeys {
		set[k] = true
	}
	if e.K8s != nil {
		for k := range e.K8s.FieldRefs {
			set[k] = true
		}
		for k := range e.K8s.ConfigMapRefs {
			set[k] = true
		}
		for k := range e.K8s.ResourceFieldRefs {
			set[k] = true
		}
		for k := range e.K8s.FileKeyRefs {
			set[k] = true
		}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
