import pytest
import requests
import yaml
import os
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_cascade_get_all():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/cascade", headers=headers, verify=False)
    assert response.status_code < 400

def test_cascade_get_foo():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/cascade/foo", headers=headers, verify=False)
    assert response.status_code < 400

def test_cascade_put_foo():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/cascades/foo.yaml', 'r', encoding='utf-8') as cascade_file:
        data = yaml.safe_load(cascade_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.put(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/cascade/foo", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_cascade_post_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/cascades/foobar.yaml', 'r', encoding='utf-8') as cascade_file:
        data = yaml.safe_load(cascade_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/cascade", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_cascade_delete_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/cascade/foobar", headers=headers, verify=False)
    assert response.status_code < 400
