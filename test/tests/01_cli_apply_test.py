import pytest
import subprocess
from config import *

def test_upload_workflows():
    workflow_success = {
        "no-group": "admin",
        "foo": "foo",
        "bar": "bar",
        "foo": "foo", # Test workflow update
    }

    workflow_failure = {
        "foo": "read-only",
        "foo": "read-only",
    }

    for workflow, profile in workflow_success.items():
        command = f"{SCAFFOLD_CLI} apply -f ../fixtures/workflows/{workflow}.yaml --profile {profile}"
        process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
        process.wait()
        assert process.returncode == 0

    for workflow, profile in workflow_failure.items():
        command = f"{SCAFFOLD_CLI} apply -f ../fixtures/workflows/{workflow}.yaml --profile {profile}"
        process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
        process.wait()
        assert process.returncode != 0
