import pytest
import subprocess
import helpers
from config import *

def do_get(profile, expected=None, cascades=None, specify_cascade=True):
    for obj in expected:
        output = subprocess.check_output([SCAFFOLD_CLI, 'get', obj, '-p', profile])
        data = helpers.loads(output.decode('utf-8'), output_json=True, padding=" ")

        assert len(data) == len(expected[obj])
        
        for idx, item in enumerate(data):
            for key in expected[obj][idx]:
                assert item[key] == expected[obj][idx][key]

    for obj in cascades:
        for cascade, num_tasks in cascades[obj].items():
            if specify_cascade:
                output = subprocess.check_output([SCAFFOLD_CLI, 'get', obj, '-p', profile, '-c', cascade])
            else:
                output = subprocess.check_output([SCAFFOLD_CLI, 'get', obj, '-p', profile])
            data = helpers.loads(output.decode('utf-8'), output_json=True, padding=" ")

            assert len(data) == num_tasks

            for item in data:
                assert item["CASCADE"] == cascade

def do_get_status(profile, cascades=None, specify_cascade=True):
    for obj in cascades:
        for cascade, status  in cascades[obj].items():
            try:
                if specify_cascade:
                    output = subprocess.check_output([SCAFFOLD_CLI, 'get', obj, '-p', profile, '-c', cascade])
                else:
                    output = subprocess.check_output([SCAFFOLD_CLI, 'get', obj, '-p', profile])
                if status == 1:
                    assert True
                else:
                    assert False
            except subprocess.CalledProcessError:
                if status == 1:
                    assert False
                else:
                    assert True

def test_admin_get():
    expected = {
        "cascade": [
            {
                "NAME": "no-group",
                "VERSION": "v1",
                "GROUPS": ""
            },
            {
                "NAME": "foo",
                "VERSION": "v1",
                "GROUPS": "foo"
            },
            {
                "NAME": "bar",
                "VERSION": "v1",
                "GROUPS": "bar"
            }
        ],
        "datastore": [
            {
                "NAME": "no-group",
            },
            {
                "NAME": "foo",
            },
            {
                "NAME": "bar",
            }
        ],
    }
    cascades = {
        "state": {
            "foo": 18,
            "bar": 18
        },
        "task": {
            "foo": 6,
            "bar": 6
        }
    }
    
    do_get('admin', expected, cascades)

def test_admin_get_individual():
    cascades = {
        "state/write_file": {
            "foo": 1,
            "bar": 1
        },
        "task/write_file": {
            "foo": 1,
            "bar": 1
        },
    }
    
    do_get_status('admin', cascades)

    cascades = {
        "cascade/foo": {
            "foo": 1,
        },
        "cascade/bar": {
            "bar": 1,
        },
        "cascade/no-group": {
            "no-group": 1,
        },
        "datastore/foo": {
            "foo": 1,
        },
        "datastore/bar": {
            "bar": 1,
        },
        "datastore/no-group": {
            "no-group": 1,
        },
    }

    do_get_status('admin', cascades, specify_cascade=False)

def test_bar_get():
    expected = {
        "cascade": [
            {
                "NAME": "no-group",
                "VERSION": "v1",
                "GROUPS": ""
            },
            {
                "NAME": "bar",
                "VERSION": "v1",
                "GROUPS": "bar"
            }
        ],
        "datastore": [
            {
                "NAME": "no-group",
            },
            {
                "NAME": "bar",
            }
        ],
    }

    cascades = {
        "state": {
            "foo": 0,
            "bar": 18
        },
        "task": {
            "foo": 0,
            "bar": 6
        }
    }
    
    do_get('bar', expected, cascades)

def test_bar_get_individual():
    cascades = {
        "state/write_file": {
            "foo": 0,
            "bar": 1,
            "no-group": 1,
        },
        "task/write_file": {
            "foo": 0,
            "bar": 1,
            "no-group": 1,
        },
    }
    
    do_get_status('bar', cascades)

    cascades = {
        "cascade/foo": {
            "foo": 0,
        },
        "cascade/bar": {
            "bar": 1,
        },
        "cascade/no-group": {
            "no-group": 1,
        },
        "datastore/foo": {
            "foo": 0,
        },
        "datastore/bar": {
            "bar": 1,
        },
        "datastore/no-group": {
            "no-group": 1,
        },
    }

    do_get_status('bar', cascades, specify_cascade=False)

def test_foo_get():
    expected = {
        "cascade": [
            {
                "NAME": "no-group",
                "VERSION": "v1",
                "GROUPS": ""
            },
            {
                "NAME": "foo",
                "VERSION": "v1",
                "GROUPS": "foo"
            }
        ],
        "datastore": [
            {
                "NAME": "no-group",
            },
            {
                "NAME": "foo",
            }
        ],
    }

    cascades = {
        "state": {
            "foo": 18,
            "bar": 0
        },
        "task": {
            "foo": 6,
            "bar": 0
        }
    }
    
    do_get('foo', expected, cascades)

def test_foo_get_individual():
    cascades = {
        "state/write_file": {
            "foo": 1,
            "bar": 0,
            "no-group": 1,
        },
        "task/write_file": {
            "foo": 1,
            "bar": 0,
            "no-group": 1,
        },
    }
    
    do_get_status('foo', cascades)

    cascades = {
        "cascade/foo": {
            "foo": 1,
        },
        "cascade/bar": {
            "bar": 0,
        },
        "cascade/no-group": {
            "no-group": 1,
        },
        "datastore/foo": {
            "foo": 1,
        },
        "datastore/bar": {
            "bar": 0,
        },
        "datastore/no-group": {
            "no-group": 1,
        },
    }

    do_get_status('foo', cascades, specify_cascade=False)

def test_read_only_get():
    expected = {
        "cascade": [
            {
                "NAME": "no-group",
                "VERSION": "v1",
                "GROUPS": ""
            },
            {
                "NAME": "foo",
                "VERSION": "v1",
                "GROUPS": "foo"
            }
        ],
        "datastore": [
            {
                "NAME": "no-group",
            },
            {
                "NAME": "foo",
            }
        ],
    }

    cascades = {
        "state": {
            "foo": 18,
            "bar": 0
        },
        "task": {
            "foo": 6,
            "bar": 0
        }
    }
    
    do_get('read-only', expected, cascades)

def test_read_only_get_individual():
    cascades = {
        "state/write_file": {
            "foo": 1,
            "bar": 0,
            "no-group": 1,
        },
        "task/write_file": {
            "foo": 1,
            "bar": 0,
            "no-group": 1,
        },
    }
    
    do_get_status('read-only', cascades)

    cascades = {
        "cascade/foo": {
            "foo": 1,
        },
        "cascade/bar": {
            "bar": 0,
        },
        "cascade/no-group": {
            "no-group": 1,
        },
        "datastore/foo": {
            "foo": 1,
        },
        "datastore/bar": {
            "bar": 0,
        },
        "datastore/no-group": {
            "no-group": 1,
        },
    }

    do_get_status('read-only', cascades, specify_cascade=False)

def test_no_group_get():
    expected = {
        "cascade": [
            {
                "NAME": "no-group",
                "VERSION": "v1",
                "GROUPS": ""
            },
        ],
        "datastore": [
            {
                "NAME": "no-group",
            },
        ],
    }

    cascades = {
        "state": {
            "foo": 0,
            "bar": 0
        },
        "task": {
            "foo": 0,
            "bar": 0
        }
    }
    
    do_get('no-group', expected, cascades)

def test_no_group_get_individual():
    cascades = {
        "state/write_file": {
            "foo": 0,
            "bar": 0,
            "no-group": 1,
        },
        "task/write_file": {
            "foo": 0,
            "bar": 0,
            "no-group": 1,
        },
    }
    
    do_get_status('no-group', cascades)

    cascades = {
        "cascade/foo": {
            "foo": 0,
        },
        "cascade/bar": {
            "bar": 0,
        },
        "cascade/no-group": {
            "no-group": 1,
        },
        "datastore/foo": {
            "foo": 0,
        },
        "datastore/bar": {
            "bar": 0,
        },
        "datastore/no-group": {
            "no-group": 1,
        },
    }

    do_get_status('no-group', cascades, specify_cascade=False)
