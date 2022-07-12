package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/pmezard/go-difflib/difflib"
)

func TestSigmafmt(t *testing.T) {
	err := filepath.Walk("testdata/sigma/rules", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		pathToRule := strings.TrimPrefix(path, "testdata/sigma/rules")
		cupaloy := cupaloy.New(
			cupaloy.SnapshotSubdirectory(filepath.Join("testdata", "snapshots", filepath.Dir(pathToRule))))

		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			output := &bytes.Buffer{}
			if err := formatPath(path, output); err != nil {
				if strings.Contains(err.Error(), "cannot unmarshal !!map into string") {
					// ignore a specific case where we don't implement the spec https://github.com/bradleyjkemp/sigma-go/issues/9
					return
				}
				t.Fatal(err)
			}
			cupaloy.SnapshotT(t, output.String())
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIdempotency(t *testing.T) {
	err := filepath.Walk("testdata/snapshots", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		pathToRule := strings.TrimPrefix(path, "testdata/snapshots/")

		t.Run(pathToRule, func(t *testing.T) {
			t.Parallel()
			input, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// Annoyingly cupaloy adds a newline to the end of all snapshots... :(
			// So to get the actual raw data back you have to trim the final \n
			input = bytes.TrimSuffix(input, []byte("\n"))

			output, _, err := formatRule(input)
			if err != nil {
				if strings.Contains(err.Error(), "cannot unmarshal !!map into string") {
					// ignore a specific case where we don't implement the spec https://github.com/bradleyjkemp/sigma-go/issues/9
					return
				}
				t.Fatal(err)
			}
			if string(input) != string(output) {
				diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(string(input)),
					FromFile: path + ".orig",
					B:        difflib.SplitLines(string(output)),
					ToFile:   path,
					Context:  2,
				})
				t.Fatal("Non idempotent, got diff:", diff)
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
