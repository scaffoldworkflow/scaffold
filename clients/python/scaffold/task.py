import json
import requests

class Task:
    keys = {
        'name': '',
        'kind': '',
        'cron': '',
        'workflow': '',
        'depends_on': {},
        'image': '',
        'run': '',
        'store': {},
        'load': {},
        'env': {},
        'inputs': {},
        'updated': '',
        'run_number': 0,
        'should_rm': False,
        'auth_execute': False,
        'disabled': False,
        'container_login_command': '',
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
    response = requests.delete(f"{base}/api/v1/task/{workflow}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def delete_individual(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.delete(f"{base}/api/v1/task/{workflow}/{name}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def get_all(base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/task", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_workflow(workflow: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/task/{workflow}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_individual(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/task/{workflow}/{name}", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def create(data: Task, base: str, auth: str, fail_on_error: bool=True) -> int:
    if not data:
        raise ValueError("No task passed in")
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.post(f"{base}/api/v1/task", headers=headers, json=data.json(), verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def update(data: Task, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.put(f"{base}/api/v1/task/{data.workflow}/{data.name}", headers=headers, json=data.json(), verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def toggle_enabled(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.put(f"{base}/api/v1/task/{workflow}/{name}/enabled", headers=headers, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()
