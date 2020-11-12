set -xe

# Pre-download modules so the `go run` output is clean
go mod download

output=$(go run github.com/bradleyjkemp/sigmafmt -l -v "$GITHUB_WORKSPACE")

if [ -z "$output" ]
then
      echo "SUCCESS: all files formatted correctly"
else
      echo "FAILURE: some files need re-formatting"
      echo "$output"
      exit 1
fi