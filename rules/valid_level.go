package rules

import "gopkg.in/yaml.v3"

var validLevel = nodeRule{"valid_level", func(node *yaml.Node) ([]Message, error) {
	var rule = struct {
		Status string `yaml:"level"`
	}{}
	if err := node.Decode(&rule); err != nil {
		return nil, err
	}

	if len(rule.Status) == 0 {
		return []Message{{Message: "All rules should have a severity level"}}, nil
	}

	switch rule.Status {
	case "low", "medium", "high", "critical":
		return nil, nil
	default:
		return []Message{{Message: "Unexpected rule level " + rule.Status}}, nil
	}
}}
