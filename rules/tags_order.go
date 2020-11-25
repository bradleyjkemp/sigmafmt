package rules

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var tagsInAlphabeticalOrder = nodeRule{"tags_alphabetical_order", func(rule *yaml.Node) ([]Message, error) {
	var tags *yaml.Node
	for i := 1; i < len(rule.Content); i += 2 {
		key, value := rule.Content[i-1], rule.Content[i]
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
