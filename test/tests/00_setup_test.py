import json
import pytest
import requests
import subprocess
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_setup_users():
    for user in ["bar", "foo", "read-only", "no-group"]:
        with open(f"../fixtures/users/{user}.json") as user_file:
            user_data = json.load(user_file)
        headers = {"Authorization" : f'X-Scaffold-API {SCAFFOLD_PRIMARY_KEY}' }
        response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/user", headers=headers, json=user_data, verify=False)
        assert response.status_code < 400

def test_setup_cli():
    user_creds = {
        "admin": "admin",
        "foo": "foo",
        "bar": "bar",
        "read-only": "read-only",
        "no-group": "no-group",
    }

    for username, password in user_creds.items():
        command = f"{SCAFFOLD_CLI} configure --username {username} --password {password} --profile {username}"
        if SCAFFOLD_PROTOCOL == "https":
            command = f"{SCAFFOLD_CLI} configure --username {username} --password {password} --profile {username} --protocol {SCAFFOLD_PROTOCOL} --skip-verify"
        process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE)
        process.wait()
        assert process.returncode == 0


