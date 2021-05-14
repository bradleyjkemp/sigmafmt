package rules

import (
	"fmt"
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
	currentGroup := 1
	var addedWhitespace bool
	for i := 0; i < len(rule.Content); i += 2 {
		key := rule.Content[i]
		canonicalGroup := canonicalKeyGroup[key.Value]
		if canonicalGroup == 0 {
			// This is a non-canonical key so we leave it in this group
			continue
		}

		switch {
		case canonicalGroup < currentGroup:
			return nil, fmt.Errorf("expected keys to already be in a sorted order")
		case canonicalGroup == currentGroup:
			continue
		case canonicalGroup > currentGroup:
			currentGroup = canonicalGroup
			// We've moved onto a new group so this key needs some whitespace above it (if it doesn't already have some)
			if strings.HasPrefix(key.HeadComment, whitespacePreservingComment) {
				continue
			}
			addedWhitespace = true
			if key.HeadComment == "" {
				key.HeadComment = whitespacePreservingComment
			} else {
				key.HeadComment = whitespacePreservingComment + "\n" + key.HeadComment
			}
		}
	}

	if addedWhitespace {
		return []Message{{Message: "Missing canonical space between YAML groups", AutoFixed: true}}, nil
	}
	return nil, nil
}}

var (
	// The order of these lines determines the "canonical" order that keys should appear in a rule and the sections they should be split into
	canonicalKeyLayout = [][]string{
		{ // Human useful metadata: what is this rule? how bad is it?
			"title",
			"description",
			"level",
			"status",
			"falsepositives",
			"references",
			"author",
			"date",
		},
		{ // The detection itself: what data does this look at? what does it match on?
			"logsource",
			"detection",
		},
		{ // Machine useful metadata (i.e. humans don't care about IDs, tags are mainly useful for indexing)
			"id",
			"tags",
		},
	}
	canonicalKeyOrdering = map[string]int{}
	canonicalKeyGroup    = map[string]int{}
)

func init() {
	var order = 1    // reserve 0 for non-canonical keys
	var groupNum = 1 // reserve 0 for non-canonical keys
	for _, group := range canonicalKeyLayout {
		for _, key := range group {
			if canonicalKeyOrdering[key] != 0 {
				panic(fmt.Sprintf("key %s already seen in the canonical key ordering. Is there a duplicate?", key))
			}
			canonicalKeyOrdering[key] = order
			canonicalKeyGroup[key] = groupNum
			order++
		}
		groupNum++
	}
}
