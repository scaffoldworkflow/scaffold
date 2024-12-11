import json
import requests

class WorkflowInput:
    keys = {
        'name': '',
        'workflow': '',
        'description': '',
        'default': '',
        'type': '',
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
    
def delete_workflow(workflow: str, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.delete(f"{base}/api/v1/input/{workflow}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def delete_name(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.delete(f"{base}/api/v1/input/{workflow}/{name}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def all(base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/input", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_workflow(workflow: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/input/{workflow}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_name(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/input/{workflow}/{name}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def create(base: str, auth: str, fail_on_error: bool=True, data: WorkflowInput=None) -> int:
    if not data:
        raise ValueError("No input passed in")
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.post(f"{base}/api/v1/input", headers=headers, json=data.json(), verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def update(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True, data: WorkflowInput=None) -> tuple[int, any]:
    if not data:
        raise ValueError("No input passed in")
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.put(f"{base}/api/v1/input/{workflow}/{name}", headers=headers, json=data.json(), verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def update_dependent_tasks(name: str, changed_inputs: list[str], base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.post(f"{base}/api/v1/input/{name}/update", headers=headers, json=changed_inputs, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()
