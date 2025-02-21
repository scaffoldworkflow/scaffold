#! /usr/bin/env bash

set -eo pipefail

if [[ "${SCAFFOLD_DEBUG}" == "true" ]]; then
    set -x
fi

# ======== COLORS ======== #

RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
NO_COLOR='\033[0m'

# ======== LOGGING ======== #

# Set log level to INFO if it doesn't exist
if [[ -z "${LOG_LEVEL}" ]]; then
    export LOG_LEVEL="INFO"
fi

# Setup variable for log level comparison
function _set_log_level {
    case ${LOG_LEVEL} in
        FATAL)
            export LOG_LEVEL_INT=0
            ;;
        SUCCESS)
            export LOG_LEVEL_INT=1
            ;;
        ERROR)
            export LOG_LEVEL_INT=2
            ;;
        WARN)
            export LOG_LEVEL_INT=3
            ;;
        INFO)
            export LOG_LEVEL_INT=4
            ;;
        DEBUG)
            export LOG_LEVEL_INT=5
            ;;
        TRACE)
            export LOG_LEVEL_INT=6
            ;;
        *)
            echo "[${YELLOW}WARN${NO_COLOR}] :: Invalid log level '${LOG_LEVEL}' -- setting level to 'INFO'"
            export LOG_LEVEL='INFO'
            export LOG_LEVEL_INT=4
            ;;
    esac
}

# Print out log statement depending on level
function _log {
    LEVEL=$1
    MESSAGE=$2
    case ${LEVEL} in
        FATAL)
            if [[ ${LOG_LEVEL_INT} -ge 0 ]]; then
                echo -e "[${RED}FATAL${NO_COLOR}] :: ${MESSAGE}"
                exit 1
            fi
            ;;
        SUCCESS)
            if [[ ${LOG_LEVEL_INT} -ge 1 ]]; then
                echo -e "[${GREEN}SUCCESS${NO_COLOR}] :: ${MESSAGE}"
            fi
            ;;
        ERROR)
            if [[ ${LOG_LEVEL_INT} -ge 2 ]]; then
                echo -e "[${RED}ERROR${NO_COLOR}] :: ${MESSAGE}"
            fi
            ;;
        WARN)
            if [[ ${LOG_LEVEL_INT} -ge 3 ]]; then
                echo -e "[${YELLOW}WARN${NO_COLOR}] :: ${MESSAGE}"
            fi
            ;;
        INFO)
            if [[ ${LOG_LEVEL_INT} -ge 4 ]]; then
                echo -e "[${GREEN}INFO${NO_COLOR}] :: ${MESSAGE}"
            fi
            ;;
        DEBUG)
            if [[ ${LOG_LEVEL_INT} -ge 5 ]]; then
                echo -e "[${CYAN}DEBUG${NO_COLOR}] :: ${MESSAGE}"
            fi
            ;;
        TRACE)
            if [[ ${LOG_LEVEL_INT} -ge 6 ]]; then
                echo -e "[${BLUE}TRACE${NO_COLOR}] :: ${MESSAGE}"
            fi
            ;;
        *)
            echo "Invalid log level: ${LEVEL}"
    esac
}

_set_log_level

# _log "INFO" "Checking for jq installation..."
# if ! command -v jq 2>&1 >/dev/null
# then
#     _log "INFO" "jq not installed, installing now..."
#     curl -s https://webinstall.dev/jq | bash
#     export PATH="${PATH}:${HOME}/.local/bin"
#     _log "INFO" "Done!"
# fi

export SCAFFOLD_CONTEXT="$( echo "${SCAFFOLD_CONTEXT_B64}" | base64 -d)"

_log "INFO" "Setting up run directories..."

# Shouldn't need these in the k8s job but helpful for local testing
rm -rf resources || true
rm -rf inputs || true
rm -rf outputs || true

mkdir -p resources
mkdir -p inputs
mkdir -p outputs
_log "SUCCESS" "Done!"

_log "INFO" "Loading resource types..."

pushd resources > /dev/null
echo "${SCAFFOLD_RESOURCE_TYPES}" > .resource_types.json
while read resource_type
do
    name=$(echo "$resource_type" | jq -r .name)
    script=$(echo "$resource_type" | jq -r .script)
    language=$(echo "$resource_type" | jq -r .language)
    requirements=$(echo "$resource_type" | jq -r .requirements)
    
    mkdir -p "${name}"
    pushd "${name}" > /dev/null

    _log "DEBUG" "Loading resource ${name}..."
    _log "TRACE" "Got language ${language} for resource ${name}"

    case "${language}" in
        bash)
            echo "${script}" | base64 -d > .main.sh
cat << EOF > .run.sh
    #! /usr/bin/env bash

    set -eo pipefail

    here=\$( cd -- "\$( dirname -- "\${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

    chmod +x "\${here}/.main.sh"

    "\${here}/.main.sh" "\$1" "\$2"

    deactivate

EOF
            chmod +x .run.sh
            ;;
        python)
            python -m venv .venv
            source .venv/bin/activate
            if [[ "${requirements}" != "" ]]; then
                _log "DEBUG" "Installing requirements for resource ${name}"
                echo "${requirements}" > requirements.txt
                pip install -r requirements.txt
            fi
            echo "${script}" | base64 -d > .main.py
cat << EOF > .run.sh
    #! /usr/bin/env bash

    set -eo pipefail

    here=\$( cd -- "\$( dirname -- "\${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

    source "\${here}/.venv/bin/activate"

    python "\${here}/.main.py" "\$1" "\$2"

    deactivate

EOF
            chmod +x .run.sh
            deactivate
            ;;
        *)
            _log "FATAL" "Resource language ${language} is invalid, valid languages are 'bash' and 'python'"
            ;;
    esac
    popd > /dev/null
    _log "INFO" "Loaded resource ${name}"
    
done < <(cat .resource_types.json | jq -c '.[]')
popd > /dev/null

_log "SUCCESS" "Done!"

# ================================
# RESOURCES
# ================================

_log "INFO" "Loading resources..."

pushd resources > /dev/null
echo "${SCAFFOLD_RESOURCES}" > .resources.json
popd > /dev/null

# ================================
# INPUTS
# ================================

_log "INFO" "Getting inputs..."

pushd inputs > /dev/null
echo "${SCAFFOLD_INPUTS}" > .input_resources.json
while read input_resource
do
    name="${input_resource}"
    resource_json="$(cat "../resources/.resources.json")"
    resource="$(echo "${resource_json}" | jq -r ".${name}")"
    kind="$(echo "${resource}" | jq -r '.kind')"
    args="$(echo "${resource}" | jq -c '.args')"

    _log "DEBUG" "Getting input ${name}..."
    _log "TRACE" "Got resource ${resource}"

    "./../resources/${kind}/.run.sh" "in" "${args}"

    _log "INFO" "Got input ${name}"

done < <(cat .input_resources.json | jq -r '.[]')

popd > /dev/null

_log "SUCCESS" "Done!"

# ================================
# SCRIPT EXECUTION
# ================================

case "${SCAFFOLD_LANGUAGE}" in
    "bash")
        cat .header.sh > .run.sh
        if [[ "${SCAFFOLD_SCRIPT}" == "" ]]; then
            cat "${SCAFFOLD_PATH}" >> .run.sh
        else
            echo "${SCAFFOLD_SCRIPT}" | base64 -d >> .run.sh
        fi
        chmod +x .run.sh
        . .run.sh
        ;;
    "python")
        cat .header.py > .run.py
        if [[ "${SCAFFOLD_SCRIPT}" == "" ]]; then
            cat "${SCAFFOLD_PATH}" >> .run.py
        else
            echo "${SCAFFOLD_SCRIPT}" | base64 -d >> .run.py
        fi
        python -m venv .venv
        source .venv/bin/activate
        if [[ -z "${SCAFFOLD_PYTHON_REQUIREMENTS}" ]]; then
            _log "WARN" "No python requirements defined"
        else
            _log "INFO" "Installing Python requirements"
            echo "${SCAFFOLD_PYTHON_REQUIREMENTS}" > .requirements.txt
            pip install -r .requirements.txt
        fi
        touch .context.json
        python .run.py
        export SCAFFOLD_CONTEXT=$(cat .context.json)
        ;;
esac

# ================================
# INPUTS
# ================================

_log "INFO" "Putting outputs..."

pushd outputs > /dev/null
echo "${SCAFFOLD_OUTPUTS}" > .output_resources.json
while read output_resource
do
    name="${output_resource}"
    resource_json="$(cat "../resources/.resources.json")"
    resource="$(echo "${resource_json}" | jq -r ".${name}")"
    kind="$(echo "${resource}" | jq -r '.kind')"
    args="$(echo "${resource}" | jq -c '.args')"

    _log "DEBUG" "Putting output ${name}..."
    _log "TRACE" "Got resource ${resource}"

    "./../resources/${kind}/.run.sh" "out" "${args}"

    _log "INFO" "Put output ${name}"

done < <(cat .output_resources.json | jq -r '.[]')

popd > /dev/null

_log "SUCCESS" "Done!"

if [[ "${SCAFFOLD_CONTEXT}" == "" ]]; then 
    echo "kernel::resource::context::{}"
else
    echo "kernel::resource::context::${SCAFFOLD_CONTEXT}"
fi
