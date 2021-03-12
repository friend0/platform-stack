#!/usr/bin/env bash
# Credits: https://gist.github.com/josh-padnick/fdae42c07e648c798fc27dec2367da21
set -e
default_tag=latest
readonly git_tag="${1:-$default_tag}"
github_repo_owner="altiscope"
github_repo_name="platform-stack"
output_path="/usr/local/bin/stack"
github_oauth_token="$GIT_TOKEN"
if [[ -z "$github_oauth_token" ]]; then
  printf "Error: GIT_TOKEN not set in the environment. Run 'export GIT_TOKEN=<your-git-access-token>' and retry.\n"
  exit 1
fi
release_asset_filename="stack_$([[ $OSTYPE == darwin* ]] && echo darwin || echo linux)_amd64"
# Get the "github tag id" of this release
if [[ "$git_tag" == "latest" ]]; then
	github_tag_id=$(curl --silent --show-error \
	                     --header "Authorization: token $github_oauth_token" \
	                     --request GET \
	                     "https://api.github.com/repos/$github_repo_owner/$github_repo_name/releases" \
	                     | jq --raw-output ".[0].assets | map(select(.name == \"$release_asset_filename\"))[0].id")

	download_url=$(curl --silent --show-error \
	                   --header "Authorization: token $github_oauth_token" \
	                   --header "Accept: application/vnd.github.v3.raw" \
	                   --location \
	                   --request GET \
	                   "https://api.github.com/repos/$github_repo_owner/$github_repo_name/releases" \
	                   | jq --raw-output ".[0].assets[0] | select(.name==\"$release_asset_filename\").url")
else
	github_tag_id=$(curl --silent --show-error \
	                     --header "Authorization: token $github_oauth_token" \
	                     --request GET \
	                     "https://api.github.com/repos/$github_repo_owner/$github_repo_name/releases" \
	                     | jq --raw-output ".[] | select(.tag_name==\"$git_tag\").id")
	download_url=$(curl --silent --show-error \
	                   --header "Authorization: token $github_oauth_token" \
	                   --header "Accept: application/vnd.github.v3.raw" \
	                   --location \
	                   --request GET \
	                   "https://api.github.com/repos/$github_repo_owner/$github_repo_name/releases/$github_tag_id" \
	                   | jq --raw-output ".assets[] | select(.name==\"$release_asset_filename\").url")
fi

# Get GitHub's S3 redirect URL
# Why not just curl's built-in "--location" option to auto-redirect? Because curl then wants to include all the original
# headers we added for the GitHub request, which makes AWS complain that we're trying strange things to authenticate.
redirect_url=$(curl --silent --show-error \
          --header "Authorization: token $github_oauth_token" \
          --header "Accept: application/octet-stream" \
          --request GET \
          --write-out "%{redirect_url}" \
          "$download_url")
# Finally download the actual binary
curl --silent --show-error \
          --header "Accept: application/octet-stream" \
          --output "$output_path" \
          --request GET \
          "$redirect_url"
if [[ "$?" -eq 0 ]]; then
  sudo chmod +x "$output_path"
else
  printf "Error: failed to install stack CLI"
  exit 1
fi
