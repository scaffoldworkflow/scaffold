import json
import requests

def get_individual(run_id: str, base: str, auth: str, fail_on_error: bool=True) -> tuple[int, any]:
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.get(f"{base}/api/v1/history/{run_id}", headers=headers, verify=False)
    if response.status_code < 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code, response.json()

class Workflow:
    keys = {
        'run_id': '',
        'states': [],
        'workflow': '',
        'created': '',
        'updated': '',
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
