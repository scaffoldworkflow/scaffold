#! /usr/bin/env bash

here=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

export SCAFFOLD_CONTEXT='{"foo":"bar"}'
: "${LOG_LEVEL:-INFO}"
: "${SCAFFOLD_DEBUG:-false}"

case "$1" in
    bash)
        export SCAFFOLD_LANGUAGE='bash'
        export SCAFFOLD_RESOURCE_TYPES="$(python get_resource_types.py)"
        export SCAFFOLD_RESOURCES='{"foobar":{"name":"foobar","kind":"file","args":{"path":"baz.txt"}}}'
        export SCAFFOLD_INPUTS='["foobar"]'
        export SCAFFOLD_OUTPUTS='["foobar"]'
        export SCAFFOLD_SCRIPT="$(cat main.sh | base64)"
        ;;
    python)
        export SCAFFOLD_LANGUAGE='python'
        export SCAFFOLD_PYTHON_REQUIREMENTS='pyyaml'
        export SCAFFOLD_RESOURCE_TYPES="$(python get_resource_types.py)"
        export SCAFFOLD_RESOURCES='{"foobar":{"name":"foobar","kind":"file","args":{"path":"baz.txt"}}}'
        export SCAFFOLD_INPUTS='["foobar"]'
        export SCAFFOLD_OUTPUTS='["foobar"]'
        export SCAFFOLD_SCRIPT="$(cat main.py | base64)"
        ;;
    clean)
        rm -rf "${here}/.venv" || true
        rm -rf "${here}/resources" || true
        rm -rf "${here}/inputs" || true
        rm -rf "${here}/outputs" || true
        rm -f "${here}/.run.*" || true
        rm -f "${here}/.context.json" || true
        rm -f "${here}/.header.py" || true
        rm -f "${here}/.header.sh" || true
        rm -f "${here}/.executor.sh" || true
        rm -f "${here}/.run.sh" || true
        rm -f "${here}/.run.py" || true
        rm -f "${here}/.requirements.txt" || true
        exit 0
        ;;
    *)
        echo "Invalid command, valid commands are 'bash', 'python', and 'clean'"
        exit 1
        ;;
esac

cp "${here}/../.header.py" .header.py
cp "${here}/../.header.sh" .header.sh
cp "${here}/../.executor.sh" .executor.sh
./.executor.sh
