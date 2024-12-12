import pytest
import requests
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
import scaffold.task
import time
import helpers
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_get_individual():
    test_id = helpers.task_setup()
    
    status, data = scaffold.task.get_individual(TASK_WORKFLOW, test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    t = scaffold.task.Task()
    t.loado(data)
    assert t.name == test_id

    helpers.task_teardown(test_id)

def test_get_all():
    test_id = helpers.task_setup()
    
    status, data = scaffold.task.get_all(TASK_WORKFLOW, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    found_task = False
    for datum in data:
        t = scaffold.task.Task()
        t.loado(datum)
        if t.name== test_id:
            found_task = True
            break
    assert found_task

    helpers.task_teardown(test_id)
