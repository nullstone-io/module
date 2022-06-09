package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// This file contains structs to create an internal representation of terraform config
//   before converting to nullstone manifest schema

type InternalTfConfig struct {
	Providers   map[string][]*InternalProvider `json:"provider"`
	DataSources InternalDataSources            `json:"data"`
	Variables   map[string][]*InternalVariable `json:"variable"`
	Outputs     map[string][]*InternalOutput   `json:"output"`
	Readme      string
}

func (m *InternalTfConfig) MergeIn(other InternalTfConfig) {
	if m.Providers == nil {
		m.Providers = map[string][]*InternalProvider{}
	}
	for name, provider := range other.Providers {
		m.Providers[name] = provider
	}

	if m.DataSources == nil {
		m.DataSources = InternalDataSources{}
	}
	m.DataSources.MergeIn(other.DataSources)

	if m.Variables == nil {
		m.Variables = map[string][]*InternalVariable{}
	}
	for name, variable := range other.Variables {
		m.Variables[name] = variable
	}

	if m.Outputs == nil {
		m.Outputs = map[string][]*InternalOutput{}
	}
	for name, output := range other.Outputs {
		m.Outputs[name] = output
	}
}

func (m *InternalTfConfig) ToManifest() Manifest {
	manifest := Manifest{
		Providers:    []string{},
		Connections:  map[string]Connection{},
		Variables:    map[string]Variable{},
		Outputs:      map[string]Output{},
		EnvVariables: map[string]EnvVariable{},
	}

	visitedProviders := map[string]bool{}
	for name, provider := range m.Providers {
		alias := ""
		if len(provider) > 0 {
			alias = provider[0].Alias
		}
		fullName := strings.TrimSuffix(fmt.Sprintf("%s.%s", name, alias), ".")
		if found, _ := visitedProviders[fullName]; found {
			continue
		}
		visitedProviders[fullName] = true
		manifest.Providers = append(manifest.Providers, fullName)
	}

	for name, variables := range m.Variables {
		for _, variable := range variables {
			varType := variable.Type
			if strings.HasPrefix(varType, "${") && strings.HasSuffix(varType, "}") {
				varType = strings.TrimSuffix(strings.TrimPrefix(varType, "${"), "}")
			}
			manifest.Variables[name] = Variable{
				Type:        varType,
				Description: variable.Description,
				Default:     variable.Default,
				Sensitive:   variable.Sensitive,
			}
		}
	}

	for name, outputs := range m.Outputs {
		for _, output := range outputs {
			if name == "env" || name == "secrets" {
				var envVars []InternalEnvVar
				err := json.Unmarshal(output.Value, &envVars)
				if err == nil {
					for _, envVar := range envVars {
						manifest.EnvVariables[envVar.Name] = EnvVariable{Sensitive: name == "secrets"}
					}
					continue
				}
			}

			outputType := "unknown"
			description := output.Description
			if strings.Contains(description, "|||") {
				tokens := strings.SplitN(description, "|||", 2)
				outputType = strings.TrimSpace(tokens[0])
				description = strings.TrimSpace(tokens[1])
			}
			manifest.Outputs[name] = Output{
				Type:        outputType,
				Description: description,
				Sensitive:   output.Sensitive,
			}
		}
	}

	// Collect nullstone connections
	for _, ds := range m.DataSources.OfType("ns_connection") {
		if val, ok := ds.Attrs["via"].(string); ok && val != "" {
			// If ns_connection is defined with "via",
			//   the connection is transitive and is satisfied "through" another workspace
			// We don't include this in our list of connections
			continue
		}

		name := ""
		contract := "*/*/*"
		connType := "unknown"
		optional := false
		if val, ok := ds.Attrs["name"].(string); ok {
			name = val
		}
		if val, ok := ds.Attrs["contract"].(string); ok {
			contract = val
		}
		if val, ok := ds.Attrs["type"].(string); ok {
			connType = val
		}
		if val, ok := ds.Attrs["optional"].(bool); ok {
			optional = val
		}
		manifest.Connections[name] = Connection{
			Contract: contract,
			Type:     connType,
			Optional: optional,
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

// { "ns_connection": { "name": [{ ...attrs }] } }
type InternalDataSources map[string]map[string][]map[string]interface{}

func (d InternalDataSources) MergeIn(other InternalDataSources) {
	for dsType, dataSources := range other {
		var curMap map[string][]map[string]interface{}
		var ok bool
		if curMap, ok = d[dsType]; !ok {
			curMap = map[string][]map[string]interface{}{}
			d[dsType] = curMap
		}
		for name, attrs := range dataSources {
			curMap[name] = attrs
		}
	}
}

func (d InternalDataSources) OfType(dataSourceType string) []InternalDataSource {
	all := make([]InternalDataSource, 0)
	for dsType, dataSources := range d {
		if dsType != dataSourceType {
			continue
		}
		for name, attrs := range dataSources {
			for _, attr := range attrs {
				all = append(all, InternalDataSource{
					Type:  dsType,
					Name:  name,
					Attrs: attr,
				})
			}
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
	Sensitive   bool        `json:"sensitive"`
}

type InternalOutput struct {
	Description string          `json:"description"`
	Sensitive   bool            `json:"sensitive"`
	Value       json.RawMessage `json:"value"`
}

type InternalEnvVar struct {
	Name string `json:"name"`
}
