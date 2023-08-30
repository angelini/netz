#!/usr/bin/env bash

set -euo pipefail

main() {
    cd "./dist"

    for dir in *; do
        echo "-- building ${dir}"
        (
            cd "${dir}"
            docker build -t "${dir}:latest" .
        )
    done
}

main "$@"