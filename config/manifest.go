package config

import (
	"database/sql/driver"
	"encoding/json"
)

type Manifest struct {
	Providers    []string               `json:"providers"`
	Connections  map[string]Connection  `json:"connections"`
	Variables    map[string]Variable    `json:"variables"`
	Outputs      map[string]Output      `json:"outputs"`
	EnvVariables map[string]EnvVariable `json:"env_variables"`
}

func (m *Manifest) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Manifest) Scan(src interface{}) error {
	data, ok := src.([]byte)
	if !ok {
		return nil
	}
	value := &Manifest{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*m = *value
	return nil
}

type Connection struct {
	Contract string `json:"contract"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

type Variable struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
	Sensitive   bool        `json:"sensitive"`
}

type Output struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Sensitive   bool   `json:"sensitive"`
}

type EnvVariable struct {
	Sensitive bool `json:"sensitive"`
}
