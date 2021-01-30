package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/bradleyjkemp/sigmafmt/rules"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	displayDiff   = flag.Bool("d", false, "display diffs instead of rewriting files")
	listAffected  = flag.Bool("l", false, "list files whose formatting differs from sigmafmt's")
	writeAffected = flag.Bool("w", false, "write result to (source) file instead of stdout")
	verbose       = flag.Bool("v", false, "write extra info (e.g. reasons) to stdout (implies -l)")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: sigmafmt [flags] [path ...]")
		flag.PrintDefaults()
	}
	flag.Parse()

	for _, path := range flag.Args() {
		if err := formatPath(path, os.Stdout); err != nil {
			log.Fatal(err)
		}
	}
}

// Takes a file/directory and recursively formats all .yaml/.yml files within it
func formatPath(root string, stdout io.Writer) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		switch {
		case info == nil:
			return nil
		case info.IsDir():
			return nil
		case path == root:
			// Always format a file if it's explicitly provided as an argument
		case filepath.Ext(info.Name()) != ".yaml" && filepath.Ext(info.Name()) != ".yml":
			return nil
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading %s: %w", path, err)
		}
		formatted, results, err := formatRule(contents)
		if err != nil {
			return fmt.Errorf("failed to format %s: %w", path, err)
		}
		if len(results) == 0 {
			// No rules found any issues
			return nil
		}

		if *writeAffected {
			if err := ioutil.WriteFile(path, formatted, 0); err != nil {
				return err
			}
		}
		if *listAffected || *verbose {
			fmt.Fprintln(stdout, path)
		}
		if *verbose {
			for _, result := range results {
				if result.AutoFixed {
					fmt.Fprintf(stdout, "\t[auto-fixed] %s\n", result.Message)
				} else {
					fmt.Fprintf(stdout, "\t[fix-needed] %s\n", result.Message)
				}
			}
		}
		if *displayDiff {
			diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(contents)),
				FromFile: path + ".orig",
				B:        difflib.SplitLines(string(formatted)),
				ToFile:   path,
				Context:  2,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(stdout, diff)
		}

		// If no flags are given, print the formatted file
		if !*writeAffected && !*listAffected && !*displayDiff {
			fmt.Fprint(stdout, string(formatted))
		}
		return nil
	})
}

// Takes the contents of a Sigma rule, runs all the linter rules, and returns the formatted output
func formatRule(contents []byte) ([]byte, []rules.Message, error) {
	var results []rules.Message
	for _, rule := range rules.Rules {
		formatted, messages, err := rule.Apply(contents)
		if err != nil {
			return nil, nil, err
		}

		contents = formatted
		results = append(results, messages...)
	}
	return contents, results, nil
}
