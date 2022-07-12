package rules

import (
	"fmt"
	"path"
	"strconv"

	"github.com/bradleyjkemp/sigma-go"
)

var undefinedSearch = sigmaRule{"undefined_search", func(rule sigma.Rule) ([]Message, error) {
	messages := []Message{}
	for _, condition := range rule.Detection.Conditions {
		searches, err := getReferencedSearches(condition.Search)
		if err != nil {
			return nil, err
		}

		for _, search := range searches {
			switch s := search.(type) {
			case sigma.SearchIdentifier:
				if _, defined := rule.Detection.Searches[s.Name]; !defined {
					messages = append(messages, Message{
						Message: fmt.Sprintf("condition references undefined search %s", strconv.Quote(s.Name)),
					})
				}
			case sigma.AllOfPattern:
				messages = append(messages, checkPattern(s.Pattern, rule)...)
			case sigma.OneOfPattern:
				messages = append(messages, checkPattern(s.Pattern, rule)...)

			default:
				return nil, fmt.Errorf("internal: unexpected search type %T", s)
			}
		}
	}
	return messages, nil
}}

func checkPattern(pattern string, rule sigma.Rule) []Message {
	for search := range rule.Detection.Searches {
		if matched, _ := path.Match(pattern, search); matched {
			return nil
		}
	}
	return []Message{{Message: fmt.Sprintf("pattern %s doesn't match any searches", strconv.Quote(pattern))}}
}

func getReferencedSearches(expr sigma.SearchExpr) ([]interface{}, error) {
	switch e := expr.(type) {
	case sigma.And:
		searches := []interface{}{}
		for _, s := range e {
			a, err := getReferencedSearches(s)
			if err != nil {
				return nil, err
			}
			searches = append(searches, a...)
		}
		return searches, nil

	case sigma.Or:
		searches := []interface{}{}
		for _, s := range e {
			a, err := getReferencedSearches(s)
			if err != nil {
				return nil, err
			}
			searches = append(searches, a...)
		}
		return searches, nil

	case sigma.Not:
		return getReferencedSearches(e.Expr)

	case sigma.SearchIdentifier:
		return []interface{}{e}, nil

	case sigma.AllOfIdentifier:
		return []interface{}{e.Ident}, nil
	case sigma.OneOfIdentifier:
		return []interface{}{e.Ident}, nil
	case sigma.AllOfPattern:
		return []interface{}{e}, nil
	case sigma.OneOfPattern:
		return []interface{}{e}, nil
	case sigma.AllOfThem, sigma.OneOfThem:
		return nil, nil

	default:
		return nil, fmt.Errorf("internal: unknown node type %T", e)
	}
}
