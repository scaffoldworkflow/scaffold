import pytest
import requests
import yaml
import os
import time
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def check_state(cascade, task, expected, token, success):
    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/{cascade}/{task}", headers=headers, verify=False)
    if success:
        assert response.status_code < 400
        state = response.json()
        print(state)
        if expected == 'running':
            assert state["status"] == expected or state["status"] == "waiting"
        else:
            assert state["status"] == expected
    else:
        assert response.status_code >= 400

def test_input_change():
    # configure request
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/datastore/foo", headers=headers, verify=False)
    assert response.status_code < 400

    datastore = response.json()
    datastore['env']['greeting'] = 'This is a new greeting'

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.put(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/datastore/foo", headers=headers, json=datastore, verify=False)
    assert response.status_code < 400

    time.sleep(4)

    check_state('foo', 'write_file', 'running', token, True)
