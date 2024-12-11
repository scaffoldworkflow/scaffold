
import json
import requests

class File:
    keys = {
        'name': '',
        'modified': '',
        'workflow': '',
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

def get_all(base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/file", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_workflow(workflow: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/file/{workflow}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def get_individual(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/file/{workflow}/{name}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

def create(data: File, base: str, auth: str, fail_on_error: bool=True) -> int:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.post(f"{base}/api/v1/file", headers=headers, json=data.json(), verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code

# def download(workflow: str, name: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
#     headers = {"Authorization" : f'X-Scaffold-API {auth}' }
#     response = requests.get(f"{base}/api/v1/file/{workflow}/{name}/download", headers=headers, verify=False)
#     if response.status_code < 400 and fail_on_error:
#         raise ValueError(f"Post request responded with {response.status_code}")
#     return response.status_code, response.json()
