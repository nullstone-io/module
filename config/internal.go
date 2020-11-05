package config

import (
	"fmt"
	"strings"

	"github.com/nullstone-io/terraform-parser/module"
)

// This file contains structs to create an internal representation of terraform config
//   before converting to nullstone manifest schema

type InternalTfConfig struct {
	Providers   map[string]*InternalProvider `json:"provider"`
	DataSources InternalDataSources          `json:"data"`
	Variables   map[string]*InternalVariable `json:"variable"`
	Outputs     map[string]*InternalOutput   `json:"output"`
}

func (m *InternalTfConfig) MergeIn(other InternalTfConfig) {
	if m.Providers == nil {
		m.Providers = map[string]*InternalProvider{}
	}
	for name, provider := range other.Providers {
		m.Providers[name] = provider
	}

	if m.Variables == nil {
		m.Variables = map[string]*InternalVariable{}
	}
	for name, variable := range other.Variables {
		m.Variables[name] = variable
	}

	if m.Outputs == nil {
		m.Outputs = map[string]*InternalOutput{}
	}
	for name, output := range other.Outputs {
		m.Outputs[name] = output
	}
}

func (m *InternalTfConfig) ToManifest() module.Manifest {
	manifest := module.Manifest{
		Providers:   []string{},
		Connections: map[string]module.Connection{},
		Variables:   map[string]module.Variable{},
		Outputs:     map[string]module.Output{},
	}

	visitedProviders := map[string]bool{}
	for name, provider := range m.Providers {
		fullName := strings.TrimSuffix(fmt.Sprintf("%s.%s", name, provider.Alias), ".")
		if found, _ := visitedProviders[fullName]; found {
			continue
		}
		visitedProviders[fullName] = true
		manifest.Providers = append(manifest.Providers, fullName)
	}

	for name, variable := range m.Variables {
		varType := variable.Type
		if strings.HasPrefix(varType, "${") && strings.HasSuffix(varType, "}") {
			varType = strings.TrimSuffix(strings.TrimPrefix(varType, "${"), "}")
		}
		manifest.Variables[name] = module.Variable{
			Type:        varType,
			Description: variable.Description,
			Default:     variable.Default,
		}
	}

	for name, output := range m.Outputs {
		manifest.Outputs[name] = module.Output{
			Type:        "unknown",
			Description: output.Description,
			Sensitive:   output.Sensitive,
		}
	}

	// Collect nullstone connections
	for _, ds := range m.DataSources.OfType("ns_connection") {
		connType := "unknown"
		if val, ok := ds.Attrs["type"].(string); ok {
			connType = val
		}
		manifest.Connections[ds.Name] = module.Connection{
			Type: connType,
		}
	}

	/*

		for name := range module.ProviderRequirements.RequiredProviders {
			if _, ok := visitedProviders[name]; !ok {
				manifest.Providers = append(manifest.Providers, name)
			}
		}
	*/

	return manifest
}

type InternalProvider struct {
	Alias string `json:"alias,omitempty"`
}

// { "ns_connection": { "name": { ...attrs } } }
type InternalDataSources map[string]map[string]map[string]interface{}

func (d InternalDataSources) OfType(dataSourceType string) []InternalDataSource {
	all := make([]InternalDataSource, 0)
	for dsType, dataSources := range d {
		if dsType != dataSourceType {
			continue
		}
		for name, attrs := range dataSources {
			all = append(all, InternalDataSource{
				Type:  dsType,
				Name:  name,
				Attrs: attrs,
			})
		}
	}
	return all
}

type InternalDataSource struct {
	Type  string
	Name  string
	Attrs map[string]interface{}
}

type InternalVariable struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
}

type InternalOutput struct {
	Value       string `json:"value"`
	Description string `json:"description"`
	Sensitive   bool   `json:"sensitive"`
}
