import pytest
import requests
from config import *
from requests.packages.urllib3.exceptions import InsecureRequestWarning
import scaffold.user
import uuid
import time
import helpers
requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

def test_create():
    test_id = str(uuid.uuid4())

    u = scaffold.user.User()
    u.loadf(USER_FIXTURE_PATH)
    u.username = test_id

    status = scaffold.user.create(u, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 201

    helpers.user_teardown(test_id)

def test_get_individual():
    test_id = helpers.user_setup()
    
    status, data = scaffold.user.get_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    u = scaffold.user.User()
    u.loado(data)
    assert u.username == test_id

    helpers.user_teardown(test_id)

def test_get_all():
    test_id = helpers.user_setup()
    
    status, data = scaffold.user.get_all(SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
    found_user = False
    for datum in data:
        u = scaffold.user.User()
        u.loado(datum)
        if u.username == test_id:
            found_user = True
            break
    assert found_user

    helpers.user_teardown(test_id)

def test_update():
    test_id = helpers.user_setup()

    _, data = scaffold.user.get_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    u = scaffold.user.User()
    u.loado(data)
    u.given_name = test_id

    status, data = scaffold.user.update(u, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200

    _, data = scaffold.user.get_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    u = scaffold.user.User()
    u.loado(data)
    assert u.given_name == test_id

    helpers.user_teardown(test_id)

def test_delete():
    test_id = helpers.user_setup()

    time.sleep(1)

    status = scaffold.user.delete_individual(test_id, SCAFFOLD_BASE, SCAFFOLD_AUTH)
    assert status == 200
