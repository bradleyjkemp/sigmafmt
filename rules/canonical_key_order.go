package rules

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

var canonicalKeyOrder = nodeRule{"canonical_key_order", func(node *yaml.Node) ([]Message, error) {
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
	var keys []*yaml.Node
	originalKeyOrder := map[*yaml.Node]int{}
	values := map[string]*yaml.Node{}
	for i := 1; i < len(rule); i += 2 {
		key, value := rule[i-1], rule[i]
		keys = append(keys, key)
		originalKeyOrder[key] = i - 1
		values[key.Value] = value
	}

	lessThan := func(i, j int) bool {
		iOrder := canonicalKeyOrdering[keys[i].Value]
		jOrder := canonicalKeyOrdering[keys[j].Value]

		switch {
		case iOrder != 0 && jOrder != 0:
			// Obvious case: both keys are canonical so return their canonical orders
			return iOrder < jOrder

		case iOrder == 0 && jOrder == 0:
			// Another obvious case: both keys are non-canonical so keep the same order
			return i < j

		// The complicated case, if either key is non-canonical we need to find go find the nearest canonical key
		default:
			var nonCanonicalIndex, canonicalIndex int
			if iOrder == 0 {
				nonCanonicalIndex = i
				canonicalIndex = j
			} else {
				nonCanonicalIndex = j
				canonicalIndex = i
			}

			var start, direction int
			var up, down = 1, -1
			if nonCanonicalIndex < canonicalIndex {
				start, direction = 0, up
			} else {
				start, direction = len(keys)-1, down
			}

			nearestCanonicalOrder := 0
			for nearestCanonical := start; nearestCanonical != canonicalIndex; nearestCanonical += direction {
				order := canonicalKeyOrdering[keys[nearestCanonical].Value]
				if order != 0 {
					nearestCanonicalOrder = order
				}
			}

			if nearestCanonicalOrder == 0 {
				// degenerate case where all the relevant keys are non-canonical
				return i < j
			}

			canonicalIndexOrder := canonicalKeyOrdering[keys[canonicalIndex].Value]
			if direction == up {
				return canonicalIndexOrder < nearestCanonicalOrder
			} else {
				return nearestCanonicalOrder < canonicalIndexOrder
			}
		}
	}

	if sort.SliceIsSorted(keys, lessThan) {
		// No change
		return nil, nil
	}

	sort.SliceStable(keys, lessThan)
	node.Content[0].Content = []*yaml.Node{}
	for _, key := range keys {
		node.Content[0].Content = append(node.Content[0].Content, key)
		node.Content[0].Content = append(node.Content[0].Content, values[key.Value])
	}

	return []Message{{Message: "YAML keys weren't in the canonical order", AutoFixed: true}}, nil
}}

// The order of these lines determines the "canonical" order that keys should appear in a rule
// (if adding a new key, make sure to also add it to the map below)
const (
	titleOrder          = iota + 1
	idOrder             = iota
	descriptionOrder    = iota
	logsourceOrder      = iota
	detectionOrder      = iota
	referencesOrder     = iota
	tagsOrder           = iota
	levelOrder          = iota
	statusOrder         = iota
	falsepositivesOrder = iota
)

var canonicalKeyOrdering = map[string]int{
	// reserve 0 for non-canonical keys
	"description":    descriptionOrder,
	"detection":      detectionOrder,
	"falsepositives": falsepositivesOrder,
	"id":             idOrder,
	"level":          levelOrder,
	"logsource":      logsourceOrder,
	"references":     referencesOrder,
	"status":         statusOrder,
	"tags":           tagsOrder,
	"title":          titleOrder,
}
