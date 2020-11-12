set -e

# Pre-download modules so the `go run` output is clean (doesn't include "downloading module X" lines)
go mod download

if [ -z "$INPUT_PATHS" ]
then
  INPUT_PATHS=$GITHUB_WORKSPACE
fi

output=$(go run github.com/bradleyjkemp/sigmafmt -l -v $INPUT_PATHS)

if [ -z "$output" ]
then
  echo "SUCCESS: all files formatted correctly"
else
  echo "FAILURE: some files need re-formatting"
  echo "${output}"
  # TODO: comment on the affected files in the PR
  exit 1
fi