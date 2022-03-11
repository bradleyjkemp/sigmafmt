package rules

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var multilineDescription = nodeRule{"multiline_description", func(node *yaml.Node) ([]Message, error) {
	var description *yaml.Node
	for i := 1; i < len(node.Content); i += 2 {
		key, value := node.Content[i-1], node.Content[i]
		if key.Value == "description" {
			description = value
			break
		}
	}
	if description == nil {
		// Rule doesn't have any description
		fmt.Println("No description!")
		return nil, nil
	}

	// Trailing space forces the YAML marshaller to use single line (quoted) mode
	lines := strings.Split(description.Value, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	description.Value = strings.Join(lines, "\n")

	// If the description is multiline, specifically tell the marshaller to output literal style
	if strings.Contains(description.Value, "\n") {
		description.Style = yaml.LiteralStyle
	}

	return nil, nil
}}
