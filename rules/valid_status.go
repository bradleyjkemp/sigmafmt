package rules

import "gopkg.in/yaml.v3"

var validStatus = nodeRule{"valid_status", func(node *yaml.Node) ([]Message, error) {
	var rule = struct {
		Status string `yaml:"status"`
	}{}
	if err := node.Decode(&rule); err != nil {
		return nil, err
	}

	if len(rule.Status) == 0 {
		return []Message{{Message: "All rules should have a status"}}, nil
	}

	switch rule.Status {
	case "experimental", "testing", "stable":
		return nil, nil
	default:
		return []Message{{Message: "Unexpected rule status " + rule.Status}}, nil
	}
}}
