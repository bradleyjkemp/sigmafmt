package rules

import (
	"gopkg.in/yaml.v3"
)

var conditionTimeframeOrder = nodeRule{"condition_timeframe_order", func(rule *yaml.Node) ([]Message, error) {
	var detection *yaml.Node
	for i := 1; i < len(rule.Content); i += 2 {
		key, value := rule.Content[i-1], rule.Content[i]
		if key.Value == "detection" {
			detection = value
			break
		}
	}
	if detection == nil {
		// Rule doesn't have a detection...
		return nil, nil
	}

	// Check that:
	//	* All the search identifiers come before the condition/timeframe keys
	//	* The timeframe key (if present) comes after the condition key
	var seenCondition, seenTimeframe, needsSorting bool
	for i := 0; i < len(detection.Content); i += 2 {
		key := detection.Content[i].Value
		switch {
		case key == "condition" && seenTimeframe:
			seenCondition = true
			needsSorting = true

		case key == "timeframe":
			seenTimeframe = true

		case seenCondition || seenTimeframe:
			// Shouldn't have any search identifiers after the condition/timeframe keys
			needsSorting = true
		}
	}
	if !needsSorting {
		return nil, nil
	}

	var conditionKey *yaml.Node
	var conditionValue *yaml.Node
	var timeframeKey *yaml.Node
	var timeframeValue *yaml.Node
	sorted := detection.Content[:0]
	for i := 0; i < len(detection.Content); i += 2 {
		key := detection.Content[i].Value
		switch key {
		case "condition":
			conditionKey = detection.Content[i]
			conditionValue = detection.Content[i+1]
		case "timeframe":
			timeframeKey = detection.Content[i]
			timeframeValue = detection.Content[i+1]
		default:
			sorted = append(sorted, detection.Content[i], detection.Content[i+1])
		}
	}

	if conditionKey != nil {
		sorted = append(sorted, conditionKey, conditionValue)
	}
	if timeframeKey != nil {
		sorted = append(sorted, timeframeKey, timeframeValue)
	}
	detection.Content = sorted
	return []Message{{
		Message:   "condition/timeframe keys weren't at the end of the detection block",
		AutoFixed: true,
	}}, nil
}}
