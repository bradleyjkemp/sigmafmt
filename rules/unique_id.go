package rules

import (
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

var seenIDs = map[string]bool{}
var seenIDsMutex = sync.Mutex{}

var uniqueID = nodeRule{"unique_id", func(node *yaml.Node) ([]Message, error) {
	var rule = struct {
		ID string `yaml:"id"`
	}{}
	if err := node.Decode(&rule); err != nil {
		return nil, err
	}

	if len(rule.ID) == 0 {
		return []Message{{Message: "All rules should have an ID"}}, nil
	}

	seenIDsMutex.Lock()
	defer seenIDsMutex.Unlock()
	if alreadySeen := seenIDs[rule.ID]; alreadySeen {
		return []Message{{Message: fmt.Sprintf("Another rule already has the ID %s. All rules should have a unique ID", rule.ID)}}, nil
	}

	seenIDs[rule.ID] = true
	return nil, nil
}}
