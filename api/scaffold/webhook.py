import json
import requests

def trigger(workflow: str, task: str, base: str, auth: str, fail_on_error: bool=True, data: object=None) -> int:
    if data == None:
        data = {}
    headers = {"Authorization" : f'X-Scaffold-API {auth}' }
    response = requests.post(f"{base}/api/v1/webhook/{workflow}/{task}", headers=headers, json=data, verify=False)
    if response.status_code >= 400 and fail_on_error:
        raise ValueError(f"Post request responded with {response.status_code}")
    return response.status_code
