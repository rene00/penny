#!/bin/bash
set -euf -o pipefail

while true; do
    poetry run flask db upgrade
    if [[ "$?" == "0" ]]; then
        break
    fi
    echo flask db upgrade failed, retrying in 5 seconds...
    sleep 5
done

poetry run honcho --app-root=/app start
