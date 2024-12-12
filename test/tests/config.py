import os

SCAFFOLD_PROTOCOL=os.getenv('SCAFFOLD_PROTOCOL')
SCAFFOLD_HOST="localhost"
SCAFFOLD_PORT=2997
SCAFFOLD_WS_PORT=8080
SCAFFOLD_BASE="http://localhost:2997"
SCAFFOLD_AUTH="MyCoolPrimaryKey12345"
SCAFFOLD_CLI="../../dist/linux/amd64/scaffold"

USER_FIXTURE_PATH = "../fixtures/users/foo.json"
WORKFLOW_FIXTURE_PATH = "../fixtures/workflows/foo.json"

TASK_WORKFLOW = "foo"
STATE_WORKFLOW = "foo"
