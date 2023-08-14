import pytest
import subprocess
from config import *

def test_upload_cascades():
    cascade_success = {
        "no-group": "admin",
        "foo": "foo",
        "bar": "bar",
        "foo": "foo", # Test cascade update
    }

    cascade_failure = {
        "foo": "read-only",
        "foo": "read-only",
    }

    for cascade, profile in cascade_success.items():
        command = f"{SCAFFOLD_CLI} apply -f ../fixtures/cascades/{cascade}.yaml --profile {profile}"
        process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
        process.wait()
        assert process.returncode == 0

    for cascade, profile in cascade_failure.items():
        command = f"{SCAFFOLD_CLI} apply -f ../fixtures/cascades/{cascade}.yaml --profile {profile}"
        process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
        process.wait()
        assert process.returncode != 0
