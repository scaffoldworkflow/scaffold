#! /usr/bin/env bash

sleep_amount="${SCAFFOLD_SLEEP:-"0"}"
echo "Sleeping for ${sleep_amount} seconds"
sleep "${sleep_amount}"

run_mode="${SCAFFOLD_RUN_MODE:normal}"

RUN_DIR=$(dirname $0)

pushd "${RUN_DIR}"

if [[ "${SCAFFOLD_TLS_ENABLED}" == "true" ]]; then
    echo "Setting up certificate"
    cert_path="/tmp/certs"
    crt_name="cert.crt"
    key_name="cert.key"
    
    # Check if we've specified a directory, otherwise use default
    if ! [ -z "${SCAFFOLD_TLS_CERT_PATH}" ]; then
        echo "Cert path being changes to ${SCAFFOLD_TLS_CERT_PATH}"
        cert_path="${SCAFFOLD_TLS_CERT_PATH}"
    fi

    # Ensure directory exists
    mkdir -p "${cert_path}"

    # Find first crt file in directory if ENV contents are not present, otherwise write those contents to the file location
    if [ -z "${SCAFFOLD_TLS_CERT_CRT}" ]; then
        pattern="${cert_path}/*.crt"
        files=( $pattern )
        if [[ "${#files[*]}" == "0" ]]; then
            echo "TLS is enabled, but no crt file exists in the ${cert_path} directory"
            exit 1
        fi
        echo "Found crt with name ${files[0]}"
        crt_name="${files[0]}"
    else
        echo "Writing crt contents to ${cert_path}/${key_name}"
        echo "${SCAFFOLD_TLS_CERT_CRT}" | base64 -d > "${cert_path}/${crt_name}"
    fi
    export SCAFFOLD_TLS_CRT_PATH="${cert_path}/${crt_name}"
    
    # Find first key file in directory if ENV contents are not present, otherwise write those contents to the file location
    if [ -z "${SCAFFOLD_TLS_CERT_KEY}" ]; then
        pattern="${cert_path}/*.key"
        files=( $pattern )
        if [[ "${#files[*]}" == "0" ]]; then
            echo "TLS is enabled, but no key file exists in the ${cert_path} directory"
            exit 1
        fi
        echo "Found key with name ${files[0]}"
        key_name="${files[0]}"
    else
        echo "Writing key contents to ${cert_path}/${key_name}"
        echo "${SCAFFOLD_TLS_CERT_KEY}" | base64 -d > "${cert_path}/${key_name}"
    fi
    export SCAFFOLD_TLS_KEY_PATH="${cert_path}/${key_name}"
fi

echo "Starting Scaffold"
if [[ "${run_mode}" == "normal" ]]; then
    ./scaffold
else 
    while true; do
        echo "Service started in coverage mode"
        ./scaffold -test.coverprofile=cover.out "$@" || exit 1;
        echo "Server restarting.."
    done
fi

popd
