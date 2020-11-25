package rules

import (
	"gopkg.in/yaml.v3"
)

var Rules = []Rule{
	// Essential internal prep steps:
	preserveWhitespace,
	yamlRoundtrip,

	// Actual useful rules
	canonicalKeyOrder,
	ruleTitle,
	validStatus,
	validLevel,
	tagsInAlphabeticalOrder,

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
	apply func(node *yaml.Node) ([]Message, error)
}

func (n nodeRule) Name() string {
	return n.name
}

func (n nodeRule) Apply(contents []byte) ([]byte, []Message, error) {
	node := &yaml.Node{}
	if err := yaml.Unmarshal(contents, node); err != nil {
		return nil, nil, err
	}

	messages, err := n.apply(node)
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
