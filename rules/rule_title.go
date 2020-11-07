package rules

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var ruleTitle = nodeRule{"rule_title_style", func(node *yaml.Node) ([]Message, error) {
	var rule = struct {
		Title string `yaml:"title"`
	}{}
	if err := node.Decode(&rule); err != nil {
		return nil, err
	}

	if len(rule.Title) == 0 {
		return []Message{{Message: "All rules should have a title"}}, nil
	}

	var results []Message
	if firstRune, _ := utf8.DecodeRuneInString(rule.Title); unicode.IsLower(firstRune) {
		results = append(results, Message{
			Message: "Rule titles should begin with an upper case letter",
		})
	}
	if lastRune, _ := utf8.DecodeLastRuneInString(rule.Title); unicode.IsPunct(lastRune) {
		results = append(results, Message{
			Message: "Rule titles should not end in punctuation",
		})
	}

	if len(rule.Title) > 50 {
		results = append(results, Message{
			Message: "Rule titles should be less than 50 characters so they can be used as an alert name",
		})
	}

	if strings.HasPrefix(rule.Title, "Detects ") || strings.HasPrefix(rule.Title, "Detect ") {
		results = append(results, Message{
			Message: "Don't use a prefix in the title like \"Detects\"",
		})
	}

	return results, nil
}}
