import pytest
import requests
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
import scaffold.state
import time
import helpers
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_get_individual():
    test_id = helpers.state_setup()
    
    status, data = scaffold.state.get_individual(STATE_WORKFLOW, test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    s = scaffold.state.State()
    s.loado(data)
    assert s.task == test_id

    helpers.state_teardown(test_id)

def test_get_all():
    test_id = helpers.state_setup()
    
    status, data = scaffold.state.get_all(STATE_WORKFLOW, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    found_state = False
    for datum in data:
        s = scaffold.state.State()
        s.loado(datum)
        if s.task== test_id:
            found_state = True
            break
    assert found_state

    helpers.state_teardown(test_id)
