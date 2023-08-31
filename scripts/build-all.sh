#!/usr/bin/env bash

set -euo pipefail

REGISTRY="k3d-netzregistry.localhost:3001"

build_dist_image() {
    local dir="${1}"

    (
        cd "${dir}"
        docker build -t "${REGISTRY}/${dir}:latest" .
        docker push "${REGISTRY}/${dir}:latest"
    )
}

build_debug_server() {
    (
        cd "./debug_server"
        docker build -t "${REGISTRY}/debug-server:latest" .
        docker push "${REGISTRY}/debug-server:latest"
    )
}

main() {
    build_debug_server

    cd "./dist"

    for dir in *; do
        echo ""
        echo "-------------------------------"
        echo "| building ${dir}"
        echo "-------------------------------"
        echo ""

        build_dist_image "${dir}"
    done
}

main "$@"