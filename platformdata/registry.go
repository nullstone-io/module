// Package platformdata defines the kinds of "platform data" that a Nullstone module can
// persist into Terraform state through the `ns_platform_data` data source.
//
// Platform data is the platform contract of a module: data consumed only by Nullstone
// (UI, API, CLI, deployment tooling) as opposed to Terraform outputs, which are the
// module-to-module contract. Arcana extracts platform data from state when a state
// version is saved and serves it by kind.
//
// Each kind has one or more versions. A version knows how to validate a raw payload and
// how to merge multiple instances found in the same state file.
package platformdata

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Schema validates and merges payloads for one (kind, version) pair.
type Schema interface {
	// Validate checks a raw payload and returns a descriptive error if it is invalid.
	Validate(raw json.RawMessage) error
	// Merge combines every instance of this kind found in a single state file into one payload.
	// Each kind decides its own rule (e.g. env: exactly one instance; metrics: concatenate).
	Merge(instances []json.RawMessage) (json.RawMessage, error)
}

// Kind is a registered platform data kind and its supported versions.
type Kind struct {
	Name     string
	Versions map[int]Schema
}

var registry = map[string]Kind{
	KindEnv: {
		Name: KindEnv,
		Versions: map[int]Schema{
			1: EnvV1Schema{},
		},
	},
}

// Kinds returns the names of every registered kind, sorted.
func Kinds() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Known reports whether the kind is registered (regardless of version).
func Known(kind string) bool {
	_, ok := registry[kind]
	return ok
}

// Lookup returns the schema for a (kind, version) pair.
// ok is false when the kind or the version is unknown.
func Lookup(kind string, version int) (Schema, bool) {
	k, ok := registry[kind]
	if !ok {
		return nil, false
	}
	s, ok := k.Versions[version]
	return s, ok
}

// LatestVersion returns the highest registered version for a kind (0 if the kind is unknown).
func LatestVersion(kind string) int {
	k, ok := registry[kind]
	if !ok {
		return 0
	}
	latest := 0
	for v := range k.Versions {
		if v > latest {
			latest = v
		}
	}
	return latest
}

// Envelope is the generic shape carried by `data "ns_platform_data"`.
type Envelope struct {
	Kind    string          `json:"kind"`
	Version int             `json:"version"`
	Data    json.RawMessage `json:"data"`
}

// ValidateEnvelope applies the forward-compatibility rule for platform data:
//   - unknown kind or unknown version: not an error, but a warning is returned so that an older
//     validator (provider/arcana) never blocks a newer module
//   - known kind+version with invalid data: an error
//   - data must always be a JSON object
func ValidateEnvelope(e Envelope) (warnings []string, err error) {
	if e.Kind == "" {
		return nil, fmt.Errorf("kind is required")
	}
	if e.Version <= 0 {
		return nil, fmt.Errorf("version must be a positive integer")
	}
	if !isJsonObject(e.Data) {
		return nil, fmt.Errorf("data must be a JSON object")
	}
	if !Known(e.Kind) {
		return []string{fmt.Sprintf("platform data kind %q is not recognized by this version of Nullstone; it will be stored without validation", e.Kind)}, nil
	}
	schema, ok := Lookup(e.Kind, e.Version)
	if !ok {
		return []string{fmt.Sprintf("platform data kind %q version %d is not recognized by this version of Nullstone (latest known: %d); it will be stored without validation", e.Kind, e.Version, LatestVersion(e.Kind))}, nil
	}
	if err := schema.Validate(e.Data); err != nil {
		return nil, fmt.Errorf("invalid %s (version %d) platform data: %w", e.Kind, e.Version, err)
	}
	return nil, nil
}

func isJsonObject(raw json.RawMessage) bool {
	var probe map[string]json.RawMessage
	if len(raw) == 0 {
		return false
	}
	return json.Unmarshal(raw, &probe) == nil && probe != nil
}
