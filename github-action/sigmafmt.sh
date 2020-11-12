set -e

# Pre-download modules so the `go run` output is clean (doesn't include "downloading module X" lines)
go mod download

if [ -z "$INPUT_PATH" ]
then
  RULES_PATH=$GITHUB_WORKSPACE
else
  RULES_PATH=$GITHUB_WORKSPACE/$INPUT_PATH
fi

output=$(go run github.com/bradleyjkemp/sigmafmt -l -v "$RULES_PATH")

if [ -z "$output" ]
then
  echo "SUCCESS: all files formatted correctly"
else
  echo "FAILURE: some files need re-formatting"
  echo "${output}"
  # TODO: comment on the affected files in the PR
  exit 1
fi