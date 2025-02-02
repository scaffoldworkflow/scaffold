#! /usr/bin/env bash

#======== SCAFFOLD WORKER HEADER FILE ========#

function set_context {
    key="$1"
    val="$2"
    
    SCAFFOLD_CONTEXT="$(echo "${SCAFFOLD_CONTEXT}" | jq -c ".${key} = \"${val}\"")"
}

function get_context {
    key="$1"
    
    echo "${SCAFFOLD_CONTEXT}" | jq -r ".${key}"
}
