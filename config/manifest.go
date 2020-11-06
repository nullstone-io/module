package config

type Manifest struct {
	Providers   []string              `json:"providers"`
	Connections map[string]Connection `json:"connections"`
	Variables   map[string]Variable   `json:"variables"`
	Outputs     map[string]Output     `json:"outputs"`
}

type Connection struct {
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

type Variable struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
}

type Output struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Sensitive   bool   `json:"sensitive"`
}
