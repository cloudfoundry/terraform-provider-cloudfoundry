#!/usr/bin/env bash
set -euo pipefail

# Install the Cloud Foundry CLI v8 binary.
# Refer: https://docs.cloudfoundry.org/cf-cli/install-go-cli.html#pkg-linux
readonly HTTPS_PROTOCOL='=https'

if [[ "${EUID}" -eq 0 ]]; then
  sudo_command=()
else
  sudo_command=(sudo)
fi

temporary_directory=$(mktemp -d)
trap 'rm -rf "${temporary_directory}"' EXIT

curl -fsSL --proto "${HTTPS_PROTOCOL}" --proto-redir "${HTTPS_PROTOCOL}" \
  "https://packages.cloudfoundry.org/stable?release=linux64-binary&version=v8&source=github" \
  | tar -xzf - -C "${temporary_directory}"

# The archive provides cf8 and a cf symlink that points to it.
"${sudo_command[@]}" mv -f "${temporary_directory}/cf8" /usr/local/bin/cf8
"${sudo_command[@]}" mv -f "${temporary_directory}/cf" /usr/local/bin/cf

# Bash completion takes effect after opening a new shell.
"${sudo_command[@]}" mkdir -p /usr/share/bash-completion/completions
"${sudo_command[@]}" curl -fsSL --proto "${HTTPS_PROTOCOL}" --proto-redir "${HTTPS_PROTOCOL}" \
  -o /usr/share/bash-completion/completions/cf8 \
  https://raw.githubusercontent.com/cloudfoundry/cli-ci/master/ci/installers/completion/cf8

# Verify installation.
cf version
