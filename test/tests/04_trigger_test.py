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

def test_trigger_write_file_foo():
    # configure request
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    check_state('foo', 'write_file', 'not_started', token, True)

    # Create run
    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/run/foo/write_file", headers=headers, verify=False)
    assert response.status_code < 400

    time.sleep(4)
    
    check_state('foo', 'write_file', 'running', token, True)
    
    exit_count = 120
    counter = 0
    status = 'running'
    while status == 'running':
        headers = {"Authorization" : f'X-Scaffold-API {token}' }
        response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/write_file", headers=headers, verify=False)
        state = response.json()
        status = state['status']
        counter += 1
        if counter == exit_count:
            assert False
        time.sleep(1)
    
    assert status == 'success'

def test_trigger_write_file_read_only():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['read-only']['apitoken']

    check_state('foo', 'write_file', 'success', token, True)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/run/foo/write_file", headers=headers, verify=False)
    assert response.status_code >= 400

def test_trigger_write_file_bar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['bar']['apitoken']

    check_state('foo', 'write_file', 'not_started', token, False)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/run/foo/write_file", headers=headers, verify=False)
    assert response.status_code >= 400

def test_trigger_write_file_no_group():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['no-group']['apitoken']

    check_state('foo', 'write_file', 'not_started', token, False)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/run/foo/write_file", headers=headers, verify=False)
    assert response.status_code >= 400

def test_trigger_auto_execute():
    # configure request
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    # Create run
    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/run/foo/store_full_message", headers=headers, verify=False)
    assert response.status_code < 400

    exit_count = 60
    counter = 0
    status = 'running'
    while status == 'running':
        headers = {"Authorization" : f'X-Scaffold-API {token}' }
        response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/store_full_message", headers=headers, verify=False)
        state = response.json()
        status = state['status']
        counter += 1
        if counter == exit_count:
            assert False
        time.sleep(1)
    
    time.sleep(5)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/print_file", headers=headers, verify=False)
    state = response.json()

    exit_count = 60
    counter = 0
    status = state['status']
    while status == 'running' or state == 'waiting':
        headers = {"Authorization" : f'X-Scaffold-API {token}' }
        response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/print_file", headers=headers, verify=False)
        state = response.json()
        status = state['status']
        counter += 1
        if counter == exit_count:
            assert False
        time.sleep(1)
    assert status == 'success'

    time.sleep(10)

    check_state('foo', 'print_always', 'success', token, True)
    check_state('foo', 'print_env', 'success', token, True)

def test_trigger_check():
    # configure request
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    exit_count = 60
    counter = 0
    status = 'success'
    while status == 'success':
        headers = {"Authorization" : f'X-Scaffold-API {token}' }
        response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/print_file", headers=headers, verify=False)
        state = response.json()
        status = state['status']
        counter += 1
        if counter == exit_count:
            assert False
        time.sleep(1)
    assert status == 'error'

    time.sleep(10)

    check_state('foo', 'print_error', 'success', token, True)

def test_trigger_propagate():
    # configure request
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/run/foo/write_file", headers=headers, verify=False)
    assert response.status_code < 400

    time.sleep(4)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/print_file", headers=headers, verify=False)
    state = response.json()
    status = state['status']
    assert status == 'not_started'
