import pytest
import requests
import yaml
import os
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_task_get_all():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task", headers=headers, verify=False)
    assert response.status_code < 400

    tasks = response.json()
    assert 'tasks' in tasks

def test_task_get_foo():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task/foo", headers=headers, verify=False)
    assert response.status_code < 400

def test_task_get_foo_write_file():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['admin']['apitoken']

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.get(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task/foo/write_file", headers=headers, verify=False)
    assert response.status_code < 400

    task = response.json()
    assert 'cascade' in task

    with open('../fixtures/tasks/foo.write_file.yaml', 'w', encoding='utf-8') as task_file:
        yaml.dump(task, task_file)

def test_task_put_foo_write_file():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/tasks/foo.write_file.yaml', 'r', encoding='utf-8') as task_file:
        data = yaml.safe_load(task_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.put(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task/foo/write_file", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_task_post_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/tasks/foobar.yaml', 'r', encoding='utf-8') as task_file:
        data = yaml.safe_load(task_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task", headers=headers, json=data, verify=False)
    assert response.status_code < 400

def test_task_delete_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/tasks/foobar.yaml', 'r', encoding='utf-8') as task_file:
        data = yaml.safe_load(task_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task/foobar", headers=headers, verify=False)
    assert response.status_code < 400

def test_task_delete_foobar_foobar():
    home = os.path.expanduser('~')
    with open(f"{home}/.scaffold/config") as config_file:
        config_data = yaml.safe_load(config_file)
    token = config_data['foo']['apitoken']

    with open('../fixtures/tasks/foobar.yaml', 'r', encoding='utf-8') as task_file:
        data = yaml.safe_load(task_file)

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.post(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task", headers=headers, json=data, verify=False)
    assert response.status_code < 400

    headers = {"Authorization" : f'X-Scaffold-API {token}' }
    response = requests.delete(f"{SCAFFOLD_PROTOCOL}://{SCAFFOLD_HOST}:{SCAFFOLD_PORT}/api/v1/task/foobar/foobar", headers=headers, verify=False)
    assert response.status_code < 400
