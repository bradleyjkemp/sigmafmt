package rules

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

var Rules = []Rule{
	// Essential internal prep steps:
	preserveWhitespace,
	yamlRoundtrip,

	// Actual useful rules
	uniqueID,
	ruleTitle,
	validStatus,
	validLevel,
	tagsInAlphabeticalOrder,
	conditionTimeframeOrder,
	canonicalKeyOrder,
	whitespaceBetweenSections,

	// Essential internal cleanup steps:
	undoPreserveWhitespace,
}

type Rule interface {
	Name() string
	Apply(contents []byte) ([]byte, []Message, error)
}

type Message struct {
	Message   string
	AutoFixed bool
}

type nodeRule struct {
	name  string
	apply func(rule *yaml.Node) ([]Message, error)
}

func (n nodeRule) Name() string {
	return n.name
}

func (n nodeRule) Apply(contents []byte) ([]byte, []Message, error) {
	node := &yaml.Node{}
	if err := yaml.Unmarshal(contents, node); err != nil {
		return nil, nil, err
	}

	switch {
	case node.Kind != yaml.DocumentNode:
		return nil, nil, fmt.Errorf("expected a document node, got a %v", node.Kind)
	case len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode:
		return nil, nil, fmt.Errorf("expected Rule to consist of a single YAML map")
	}

	rule := node.Content[0].Content
	if len(rule)%2 != 0 {
		return nil, nil, fmt.Errorf("internal, please report! expected an even number of elements in a mapping node")
	}
	messages, err := n.apply(node.Content[0])
	if err != nil {
		return nil, nil, err
	}
	result, err := yaml.Marshal(node)
	return result, messages, err
}

type rawRule struct {
	name  string
	apply func(contents []byte) ([]byte, []Message, error)
}

func (r rawRule) Name() string {
	return r.name
}

func (r rawRule) Apply(contents []byte) ([]byte, []Message, error) {
	return r.apply(contents)
}
