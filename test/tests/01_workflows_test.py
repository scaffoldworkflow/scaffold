import pytest
import requests
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
import scaffold.workflow
import time
import helpers
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_create():
    test_id = helpers.user_setup()

    w = scaffold.workflow.Workflow()
    w.loadf(WORKFLOW_FIXTURE_PATH)
    w.name = test_id

    status = scaffold.workflow.create(w, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 201

    helpers.workflow_teardown(test_id)

def test_get_individual():
    test_id = helpers.workflow_setup()
    
    status, data = scaffold.workflow.get_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    w = scaffold.workflow.Workflow()
    w.loado(data)
    assert w.name == test_id

    helpers.workflow_teardown(test_id)

def test_get_all():
    test_id = helpers.workflow_setup()
    
    status, data = scaffold.workflow.get_all(SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    found_workflow = False
    for datum in data:
        w = scaffold.workflow.Workflow()
        w.loado(datum)
        if w.name== test_id:
            found_workflow = True
            break
    assert found_workflow

    helpers.workflow_teardown(test_id)

def test_update():
    test_id = helpers.workflow_setup()

    _, data = scaffold.workflow.get_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    w = scaffold.workflow.Workflow()
    w.loado(data)
    w.version = test_id

    status, data = scaffold.workflow.update(w, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200

    _, data = scaffold.workflow.get_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    w = scaffold.workflow.Workflow()
    w.loado(data)
    assert w.version == test_id

    helpers.workflow_teardown(test_id)

def test_delete():
    test_id = helpers.workflow_setup()

    time.sleep(1)

    status = scaffold.workflow.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
