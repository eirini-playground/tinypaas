#!/bin/bash

set -ex

readonly REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

pushd "$REPO_ROOT/cli" || exit 1
{
  go build -o bin/tinypaas
}
popd || exit 1

sudo ln -sfn "$REPO_ROOT/cli/bin/tinypaas" /usr/local/bin/tinypaas
