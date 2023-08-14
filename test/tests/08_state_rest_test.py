import pytest
import requests
import yaml
import os
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_state_get_all():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state", headers=headers, verify=False)
    assert response.status_code < 400

    states = response.json()
    assert 'states' in states

def test_state_get_foo():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo", headers=headers, verify=False)
    assert response.status_code < 400

def test_state_get_foo_write_file():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/write_file", headers=headers, verify=False)
    assert response.status_code < 400

    state = response.json()
    assert 'cascade' in state

    with open('../fixtures/states/foo.write_file.yaml', 'w', encoding='utf-8') as state_file:
        yaml.dump(state, state_file)

def test_state_put_foo_write_file():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/states/foo.write_file.yaml', 'r', encoding='utf-8') as state_file:
        data = yaml.safe_load(state_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.put(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foo/write_file", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_state_post_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/states/foobar.yaml', 'r', encoding='utf-8') as state_file:
        data = yaml.safe_load(state_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_state_delete_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/states/foobar.yaml', 'r', encoding='utf-8') as state_file:
        data = yaml.safe_load(state_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foobar", headers=headers, verify=False)
    assert response.status_code < 400

def test_state_delete_foobar_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/states/foobar.yaml', 'r', encoding='utf-8') as state_file:
        data = yaml.safe_load(state_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state", headers=headers, json=data, verify=False)
    assert response.status_code < 400

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/state/foobar/foobar", headers=headers, verify=False)
    assert response.status_code < 400
