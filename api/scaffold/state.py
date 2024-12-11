import json
import requests

class State:
    keys = {
        'task': '',
        'workflow': '',
        'status': '',
        'started': '',
        'finished': '',
        'output': '',
        'output_checksum': '',
        'display': [],
        'worker': '',
        'number': 0,
        'disabled': False,
        'killed': False,
        'pid': 0,
        'history': [],
        'context': {},
    }
    def __init__(self):
        for key, val in self.keys.items():
            setattr(self, key, val)

    def loadf(self, path: str) -> None:
        with open(path, 'r', encoding='utf-8') as load_file:
            data = json.load(load_file)
        for key, _ in self.keys.items():
            setattr(self, key, data.get(key, self.keys[key]))

    def loads(self, data_str: str) -> None:
        data = json.loads(data_str)
        for key, _ in self.keys.items():
            setattr(self, key, data.get(key, self.keys[key]))

    def loado(self, data: object) -> None:
        for key, _ in self.keys.items():
            setattr(self, key, data.get(key, self.keys[key]))

    def json(self) -> dict:
        out = {}
        for key, _ in self.keys.items():
            out[key] = getattr(self, key)
        return out

def delete_workflow(name: str, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.delete(f"{base}/api/v1/state/{name}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def delete_individual(workflow: str, task: str, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.delete(f"{base}/api/v1/state/{workflow}/{task}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def get_all(base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/state", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_workflow(workflow: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/state/{workflow}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_individual(workflow: str, task: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/state/{workflow}/{task}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def create(data: State, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.post(f"{base}/api/v1/state", headers=headers, json=data.json(), verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def update(data: State, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.put(f"{base}/api/v1/state/{data.workflow}/{data.task}", headers=headers, json=data.json(), verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

