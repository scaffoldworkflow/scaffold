import json
import base64

with open('../../manager/resource/file.py', 'r') as f:
    script = f.read()

out = {
    "foo": {
        "name": "file",
        "script": base64.b64encode(script.encode('utf-8')).decode('utf-8'),
        "language": "python",
        "requirements": ""
    }
}

print(json.dumps(out))
