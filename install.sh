#!/usr/bin/env bash

set -e

# Default tag = latest
readonly git_tag="${1:-latest}"
github_repo="altiscope/platform-stack"
output_path="./stack"
github_api=api.github.com
github_oauth_token="$GIT_TOKEN"

if [[ -z "$github_oauth_token" ]]; then
  printf "Error: GIT_TOKEN not set in the environment. Run 'export GIT_TOKEN=<your-git-access-token>' and retry.\n"
  exit 1
fi

release_asset_filename="stack_$([[ $OSTYPE == darwin* ]] && echo darwin || echo linux)_amd64"

if [ "$git_tag" = "latest" ]; then
  # Github should return the latest release first.
  parser=".[0].assets | map(select(.name == \"$release_asset_filename\"))[0].id"
else
  parser=". | map(select(.tag_name == \"git_tag\"))[0].assets | map(select(.name == \"$release_asset_filename\"))[0].id"
fi

token_check=`curl --show-error -o /dev/null -w "%{http_code}" -sL -H "Authorization: token $github_oauth_token" -H "Accept: application/vnd.github.v3.raw" https://$github_api/repos/$github_repo/releases`
if [[ $token_check -ne 200 ]]; then
  printf "error: failed to install stack CLI: status code %s\n" $token_check
  exit 1
fi

assets=`curl -sL -H "Authorization: token $github_oauth_token" -H "Accept: application/vnd.github.v3.raw" https://$github_api/repos/$github_repo/releases`
asset_id=`echo $assets | jq "$parser"`

curl --show-error --header 'Accept: application/octet-stream' --location --output "$output_path" --request GET \
https://$github_oauth_token:@$github_api/repos/$github_repo/releases/assets/$asset_id?access_token=$github_oauth_token

if [[ "$?" -ne 0 ]]; then
  printf "error: failed to install stack CLI"
  exit 1
fi