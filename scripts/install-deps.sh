#!/bin/bash

set -euf -o pipefail

test -f "./penny.py" || exit

BOOTSTRAP_TABLE_VERSION=1.15.5
wget -O app/static/js/bootstrap-table.js https://unpkg.com/bootstrap-table@${BOOTSTRAP_TABLE_VERSION}/dist/bootstrap-table.js

