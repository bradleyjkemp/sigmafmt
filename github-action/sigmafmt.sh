set -ex

# Pre-download modules so the `go run` output is clean (doesn't include "downloading module X" lines)
go mod download

if [ -z "$INPUT_PATH" ]
then
  RULES_PATH=$GITHUB_WORKSPACE
else
  RULES_PATH=$GITHUB_WORKSPACE/$INPUT_PATH
fi
echo "Linting path $RULES_PATH"
output=$(go run github.com/bradleyjkemp/sigmafmt -l -v "$RULES_PATH")

if [ -z "$output" ]
then
  echo "SUCCESS: all files formatted correctly"
  exit 0
fi

echo "FAILURE: some files need re-formatting"
echo "${output}"

if [ -z "$GITHUB_TOKEN" ]
then
  # No GitHub token, so can't leave any comments
  exit 1
fi

# If this is a PR, leave a comment
PR_NUMBER=$(jq --raw-output .pull_request.number "$GITHUB_EVENT_PATH")

if [ "$PR_NUMBER" = "null" ]
then
  exit 1
fi

COMMENTS=$(curl \
  -H "Accept: application/vnd.github.v3+json" \
  -H "authorization: Bearer ${GITHUB_TOKEN}" \
  "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/pulls/${PR_NUMBER}/comments")

EXISTING_COMMENT_ID=$(echo "$COMMENTS" | jq '[.[]|select(.user.login=="bradleyjkemp")][0].id')

if [ "$EXISTING_COMMENT_ID" = "null" ]
then
  curl \
    -X POST \
    -H "Accept: application/vnd.github.v3+json" \
    -H "authorization: Bearer ${GITHUB_TOKEN}" \
    "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/pulls/${PR_NUMBER}/comments" \
    -d "{\"body\": $(echo "$output" | jq -Rs .) }"
else
  curl \
    -X PATCH \
    -H "Accept: application/vnd.github.v3+json" \
    -H "authorization: Bearer ${GITHUB_TOKEN}" \
    "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/pulls/comments/${EXISTING_COMMENT_ID}" \
    -d "{\"body\": $(echo "$output" | jq -Rs .) }"
fi

exit 1
