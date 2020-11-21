package rules

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

const whitespacePreservingComment = `# magic whitespace preserving comment: github.com/bradleyjkemp/sigmafmt`

var preserveWhitespace = rawRule{"essential.preserve_whitespace", func(contents []byte) ([]byte, []Message, error) {
	// First get rid of any multiple trailing newlines in the file
	contents = bytes.TrimRight(contents, "\n")

	// Then split into lines in order to preserve empty ones
	lines := bytes.Split(contents, []byte("\n"))
	for i, line := range lines {
		if i == len(lines)-1 {
			// Don't preserve an empty last line because the yaml library always inserts one anyway
			// (and so trying to preserve it means we end up with double trailing newlines)
			continue
		}

		if len(bytes.TrimSpace(line)) == 0 {
			lines[i] = []byte(whitespacePreservingComment)
		}
	}
	return bytes.Join(lines, []byte("\n")), nil, nil
}}

var yamlRoundtrip = rawRule{"essential.yaml_roundtrip", func(contents []byte) ([]byte, []Message, error) {
	node := &yaml.Node{}
	if err := yaml.Unmarshal(contents, node); err != nil {
		return nil, nil, err
	}

	result, err := yaml.Marshal(node)
	if string(result) != string(contents) {
		return result, []Message{{"Non-standard YAML formatting", true}}, nil
	}
	return result, nil, err
}}

var undoPreserveWhitespace = rawRule{"essential.undo_preserve_whitespace", func(contents []byte) ([]byte, []Message, error) {
	lines := bytes.Split(contents, []byte("\n"))
	for i, line := range lines {
		if bytes.Contains(line, []byte(whitespacePreservingComment)) {
			lines[i] = nil
		}
	}

	// Fix any double new-line at the end of the file
	if len(lines[len(lines)-1]) == 0 && len(lines[len(lines)-2]) == 0 {
		lines = lines[:len(lines)-1]
	}

	return bytes.Join(lines, []byte("\n")), nil, nil
}}
