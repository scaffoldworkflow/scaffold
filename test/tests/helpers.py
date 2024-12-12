from typing import List, Dict, Union
import scaffold.user, scaffold.workflow
from config import *
import uuid

def user_setup() -> str:
    test_id = str(uuid.uuid4())

    u = scaffold.user.User()
    u.loadf(USER_FIXTURE_PATH)
    u.username = test_id

    scaffold.user.create(u, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    return test_id

def user_teardown(test_id: str) -> None:
    scaffold.user.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

def workflow_setup() -> str:
    test_id = str(uuid.uuid4())

    u = scaffold.user.User()
    u.loadf(USER_FIXTURE_PATH)
    u.username = test_id

    scaffold.user.create(u, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    w = scaffold.workflow.Workflow()
    w.loadf(WORKFLOW_FIXTURE_PATH)
    w.name = test_id

    scaffold.workflow.create(w, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    return test_id

def workflow_teardown(test_id: str) -> None:
    scaffold.workflow.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    scaffold.user.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

def task_setup() -> str:
    test_id = str(uuid.uuid4())

    u = scaffold.user.User()
    u.loadf(USER_FIXTURE_PATH)
    u.username = test_id

    scaffold.user.create(u, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    w = scaffold.workflow.Workflow()
    w.loadf(WORKFLOW_FIXTURE_PATH)
    w.name = test_id
    w.tasks[0]['name'] = test_id

    scaffold.workflow.create(w, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    return test_id

def task_teardown(test_id: str) -> None:
    scaffold.workflow.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    scaffold.user.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

def state_setup() -> str:
    test_id = str(uuid.uuid4())

    u = scaffold.user.User()
    u.loadf(USER_FIXTURE_PATH)
    u.username = test_id

    scaffold.user.create(u, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    w = scaffold.workflow.Workflow()
    w.loadf(WORKFLOW_FIXTURE_PATH)
    w.name = test_id
    w.tasks[0]['name'] = test_id

    scaffold.workflow.create(w, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    return test_id

def state_teardown(test_id: str) -> None:
    scaffold.workflow.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)

    scaffold.user.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
