package rules

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var tagsInAlphabeticalOrder = nodeRule{"tags_alphabetical_order", func(node *yaml.Node) ([]Message, error) {
	switch {
	case node.Kind != yaml.DocumentNode:
		return nil, fmt.Errorf("expected a document node, got a %v", node.Kind)
	case len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode:
		return nil, fmt.Errorf("expected Rule to consist of a single YAML map")
	}

	rule := node.Content[0].Content
	if len(rule)%2 != 0 {
		return nil, fmt.Errorf("internal, please report! expected an even number of elements in a mapping node")
	}

	var tags *yaml.Node
	for i := 1; i < len(rule); i += 2 {
		key, value := rule[i-1], rule[i]
		if key.Value == "tags" {
			tags = value
			break
		}
	}
	if tags == nil {
		// Rule doesn't have any tags
		return nil, nil
	}

	tagComparator := func(i, j int) bool {
		iParts := strings.Split(tags.Content[i].Value, ".")
		jParts := strings.Split(tags.Content[j].Value, ".")

		numParts := len(iParts)
		if len(jParts) < numParts {
			numParts = len(jParts)
		}

		for p := 0; p < numParts; p++ {
			if iParts[p] == jParts[p] {
				continue
			}

			return iParts[p] < jParts[p]
		}

		// All the shared prefix parts of the tag are equal so sort based on number of parts
		return len(iParts) < len(jParts)
	}
	if sort.SliceIsSorted(tags.Content, tagComparator) {
		return nil, nil
	}

	sort.SliceStable(tags.Content, tagComparator)
	return []Message{{
		Message:   "Tags aren't sorted alphabetically",
		AutoFixed: true,
	}}, nil
}}
