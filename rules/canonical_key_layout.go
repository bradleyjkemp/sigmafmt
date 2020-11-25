package rules

import (
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var canonicalKeyOrder = nodeRule{"canonical_key_order", func(rule *yaml.Node) ([]Message, error) {
	var keys []*yaml.Node
	originalKeyOrder := map[*yaml.Node]int{}
	values := map[string]*yaml.Node{}
	for i := 1; i < len(rule.Content); i += 2 {
		key, value := rule.Content[i-1], rule.Content[i]
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
	rule.Content = []*yaml.Node{}
	for _, key := range keys {
		rule.Content = append(rule.Content, key)
		rule.Content = append(rule.Content, values[key.Value])
	}

	return []Message{{Message: "YAML keys weren't in the canonical order", AutoFixed: true}}, nil
}}

// Ensures there's whitespace between the groups of keys (as defined below)
var whitespaceBetweenSections = nodeRule{"whitespace_between_sections", func(rule *yaml.Node) ([]Message, error) {
	// We start off looking for the first key in the first group
	currentGroup := 0
	currentKey := 0
	keyNeedsWhitespace := false
	for i := 0; i < len(rule.Content); i += 2 {
		key := rule.Content[i]
		if keyNeedsWhitespace {
			// Add a single whitespace line before this node (if it doesn't already have one)
			if !strings.HasPrefix(key.HeadComment, whitespacePreservingComment) {
				if key.HeadComment == "" {
					key.HeadComment = whitespacePreservingComment
				} else {
					key.HeadComment = whitespacePreservingComment + "\n" + key.HeadComment
				}
			}
			keyNeedsWhitespace = false
		}
		if key.Value != canonicalKeyLayout[currentGroup][currentKey] {
			// This is a non-canonical key so we leave it in this group
			continue
		}

		// This key matches the one we're looking for so move our pointer on to the next one
		currentKey++

		if currentKey >= len(canonicalKeyLayout[currentGroup]) {
			// This is the last canonical key in the group
			currentGroup++
			currentKey = 0
			// The next key needs whitespace before it
			keyNeedsWhitespace = true
		}
		if currentGroup >= len(canonicalKeyLayout) {
			// We've gone through all the canonical keys we expect so we're done
			break
		}
	}

	return []Message{{Message: "YAML keys weren't in the canonical order", AutoFixed: true}}, nil
}}

var (
	// The order of these lines determines the "canonical" order that keys should appear in a rule and the sections they should be split into
	canonicalKeyLayout = [][]string{
		{
			"title",
			"id",
			"description",
		},
		{
			"logsource",
			"detection",
		},
		{
			"references",
			"tags",
			"level",
			"status",
			"falsepositives",
		},
	}
	canonicalKeyOrdering = map[string]int{}
)

func init() {
	var order = 1 // reserve 0 for non-canonical keys
	for _, group := range canonicalKeyLayout {
		for _, key := range group {
			canonicalKeyOrdering[key] = order
			order++
		}
	}
}
