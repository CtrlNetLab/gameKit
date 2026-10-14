#!/bin/bash

usage() {
    cat <<EOF
Usage:
    ${this} service

EOF
    exit 0
}

main() {
    this="$(readlink -f "$0")"
    if [ "$#" -lt 1 ]; then
        usage
    fi

    service="$1"
    shift

    DIR="$(dirname "${this}")"
    cd "${DIR}" || return

    control="${DIR}/control/bin/control"
    yaml="${DIR}/${service}/conf/service.yaml"
    host="$(sed -E -n '/^Control:/,/^[A-Za-z]/ s#^[ \t]+Service:[ \t]+"([^"]*)".*#\1#p' "${yaml}")"

    "${control}" "${host}" "$@"
}

main "$@"
