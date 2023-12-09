#!/bin/bash
set -eo pipefail
jq -r '.[] | "\(.name) --force --color=\(.color) --description=\"\(.description)\""' \
    .github/labels.json | xargs -L1 gh label create
