import pytest
import requests
import yaml
import os
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_datastore_get_all():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/datastore", headers=headers, verify=False)
    assert response.status_code < 400

def test_datastore_get_foo():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/datastore/foo", headers=headers, verify=False)
    assert response.status_code < 400

    datastore = response.json()
    with open('../fixtures/datastores/foo.yaml', 'w', encoding='utf-8') as datastore_file:
        yaml.dump(datastore, datastore_file)

def test_datastore_put_foo():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    with open('../fixtures/datastores/foo.yaml', 'r', encoding='utf-8') as datastore_file:
        data = yaml.safe_load(datastore_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.put(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/datastore/foo", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_datastore_delete_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    with open('../fixtures/cascades/foobar.yaml', 'r', encoding='utf-8') as cascade_file:
        data = yaml.safe_load(cascade_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/cascade", headers=headers, json=data, verify=False)
    assert response.status_code < 400

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/datastore/foobar", headers=headers, verify=False)
    assert response.status_code < 400

def test_datastore_post_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    with open('../fixtures/datastores/foobar.yaml', 'r', encoding='utf-8') as datastore_file:
        data = yaml.safe_load(datastore_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/datastore", headers=headers, json=data, verify=False)
    assert response.status_code < 400


