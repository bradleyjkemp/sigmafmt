package rules

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var ruleTitle = nodeRule{"rule_title_style", func(node *yaml.Node) ([]Message, error) {
	var title *yaml.Node
	for i := 1; i < len(node.Content); i += 2 {
		key, value := node.Content[i-1], node.Content[i]
		if key.Value == "title" {
			title = value
			break
		}
	}
	if title == nil || len(title.Value) == 0 {
		// Rule doesn't have any title
		return []Message{{Message: "All rules should have a title"}}, nil
	}

	var results []Message
	if firstRune, runeLen := utf8.DecodeRuneInString(title.Value); unicode.IsLower(firstRune) {
		title.Value = string(unicode.ToUpper(firstRune)) + title.Value[runeLen:]
		results = append(results, Message{
			AutoFixed: true,
			Message:   "Rule titles should begin with an upper case letter",
		})
	}
	lastRune, runeLen := utf8.DecodeLastRuneInString(title.Value)
	switch lastRune {
	case '.', '!', '?':
		title.Value = title.Value[:len(title.Value)-runeLen]
		results = append(results, Message{
			AutoFixed: true,
			Message:   "Rule titles should not end in punctuation",
		})
	}

	if len(title.Value) > 120 {
		results = append(results, Message{
			Message: "Rule titles should be less than 120 characters so they can be used as an alert name. Leave explanation to the description field",
		})
	}

	if strings.HasPrefix(title.Value, "Detects ") || strings.HasPrefix(title.Value, "Detect ") {
		results = append(results, Message{
			Message: "Don't use a prefix in the title like \"Detects\"",
		})
	}

	return results, nil
}}
