set -ex

if [ -z "$INPUT_PATH" ]
then
  RULES_PATH=$GITHUB_WORKSPACE
else
  RULES_PATH=$GITHUB_WORKSPACE/$INPUT_PATH
fi

_main() {
  if _dont_have_github_token; then
    _run_sigmafmt
    if _sigmafmt_found_issues; then
      echo "FAILURE: some files need re-formatting"
      echo "${SIGMAFMT_OUTPUT}"
      exit 1
    else
      echo "SUCCESS: all files formatted correctly"
      exit 0
    fi
  fi

  _run_sigmafmt_autofix
  if _git_is_dirty
  then
    (
      cd "$GITHUB_WORKSPACE"
      git add $RULES_PATH
      git commit --message "sigmafmt auto-fixed"
      git push origin
    )
  fi

  # Now run Sigma again to check for any un-fixable errors and add a PR comment
  _run_sigmafmt
  if _sigmafmt_found_issues; then
    echo "FAILURE: some files need re-formatting"
    echo ${SIGMAFMT_OUTPUT}
    _upsert_pr_comment
    exit 1
  else
    echo "SUCCESS: all files formatted correctly"
    exit 0
  fi
}

_dont_have_github_token() {
  [ -z "$GITHUB_TOKEN" ]
}

_sigmafmt_found_issues() {
  [ -n "$SIGMAFMT_OUTPUT" ]
}

_run_sigmafmt() {
  # Pre-download modules so the `go run` output is clean (doesn't include "downloading module X" lines)
  go mod download

  if [ -z "$INPUT_PATH" ]
  then
    RULES_PATH=$GITHUB_WORKSPACE
  else
    RULES_PATH=$GITHUB_WORKSPACE/$INPUT_PATH
  fi
  echo "Linting path $RULES_PATH"
  SIGMAFMT_OUTPUT=$(go run github.com/bradleyjkemp/sigmafmt -l -v "$RULES_PATH")
}

_run_sigmafmt_autofix() {
  # Pre-download modules so the `go run` output is clean (doesn't include "downloading module X" lines)
  go mod download

  if [ -z "$INPUT_PATH" ]
  then
    RULES_PATH=$GITHUB_WORKSPACE
  else
    RULES_PATH=$GITHUB_WORKSPACE/$INPUT_PATH
  fi
  echo "Linting path $RULES_PATH"
  SIGMAFMT_OUTPUT=$(go run github.com/bradleyjkemp/sigmafmt -w "$RULES_PATH")
}

_git_is_dirty() {
    [ -n "$(cd "$GITHUB_WORKSPACE" && git status -s -- "$INPUT_PATH")" ]
}

_upsert_pr_comment() {
  # If this is a PR, leave a comment
  PR_NUMBER=$(jq --raw-output .pull_request.number "$GITHUB_EVENT_PATH")

  if [ "$PR_NUMBER" = "null" ]
  then
    exit 1
  fi

  COMMENTS=$(curl \
    -H "Accept: application/vnd.github.v3+json" \
    -H "authorization: Bearer ${GITHUB_TOKEN}" \
    "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments")

  EXISTING_COMMENT_ID=$(echo "$COMMENTS" | jq '[.[]|select(.user.login=="github-actions[bot]")][0].id')

  if [ "$EXISTING_COMMENT_ID" = "null" ]
  then
    curl \
      -X POST \
      -H "Accept: application/vnd.github.v3+json" \
      -H "authorization: Bearer ${GITHUB_TOKEN}" \
      "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
      -d "{\"body\": $(echo "$SIGMAFMT_OUTPUT" | jq -Rs .) }"
  else
    curl \
      -X PATCH \
      -H "Accept: application/vnd.github.v3+json" \
      -H "authorization: Bearer ${GITHUB_TOKEN}" \
      "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/comments/${EXISTING_COMMENT_ID}" \
      -d "{\"body\": $(echo "$SIGMAFMT_OUTPUT" | jq -Rs .) }"
  fi
  exit 1
}

_remove_pr_comment() {
  # If this is a PR, leave a comment
  PR_NUMBER=$(jq --raw-output .pull_request.number "$GITHUB_EVENT_PATH")

  if [ "$PR_NUMBER" = "null" ]
  then
    exit 1
  fi

  COMMENTS=$(curl \
    -H "Accept: application/vnd.github.v3+json" \
    -H "authorization: Bearer ${GITHUB_TOKEN}" \
    "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments")

  EXISTING_COMMENT_ID=$(echo "$COMMENTS" | jq '[.[]|select(.user.login=="github-actions[bot]")][0].id')

  if [ "$EXISTING_COMMENT_ID" != "null" ]
  then
    curl \
      -X DELETE \
      -H "Accept: application/vnd.github.v3+json" \
      -H "authorization: Bearer ${GITHUB_TOKEN}" \
      "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/comments/${EXISTING_COMMENT_ID}"
  fi
  exit 1
}

_main