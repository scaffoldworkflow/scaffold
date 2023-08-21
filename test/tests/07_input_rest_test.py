import pytest
import requests
import yaml
import os
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_input_get_all():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input", headers=headers, verify=False)
    assert response.status_code < 400

def test_input_get_foo():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input/foo", headers=headers, verify=False)
    assert response.status_code < 400

def test_input_get_foo_name():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input/foo/name", headers=headers, verify=False)
    assert response.status_code < 400

    input = response.json()
    with open('../fixtures/inputs/foo.name.yaml', 'w', encoding='utf-8') as input_file:
        yaml.dump(input, input_file)

def test_input_put_foo_write_file():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    with open('../fixtures/inputs/foo.name.yaml', 'r', encoding='utf-8') as input_file:
        data = yaml.safe_load(input_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.put(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input/foo/name", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_input_post_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/inputs/foobar.yaml', 'r', encoding='utf-8') as input_file:
        data = yaml.safe_load(input_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_input_delete_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/inputs/foobar.yaml', 'r', encoding='utf-8') as input_file:
        data = yaml.safe_load(input_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input/foobar", headers=headers, verify=False)
    assert response.status_code < 400

def test_input_delete_foobar_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/inputs/foobar.yaml', 'r', encoding='utf-8') as input_file:
        data = yaml.safe_load(input_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input", headers=headers, json=data, verify=False)
    assert response.status_code < 400

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/input/foobar/foobar", headers=headers, verify=False)
    assert response.status_code < 400
