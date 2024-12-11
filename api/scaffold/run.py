import json
import requests

class Run:
    keys = {
        'name': '',
        'task': {},
        'state': {},
        'number': 0,
        'groups': [],
        'worker': '',
        'pid': 0,
        'context': {},
        'run_id': '',
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

def kill(workflow: str, task: str, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.delete(f"{base}/api/v1/run/{workflow}/{task}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

def status(run_id: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/run/{run_id}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def create(data: Run, workflow: str, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.post(f"{base}/api/v1/run/{workflow}/{data.task}", headers=headers, json=data.json(), verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code


