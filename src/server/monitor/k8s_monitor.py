import base64
import random
import time
import re
import sys
import json

# ======== MODIFIED HEADER FILE FROM CALLIGRAPHY SCRIPTING ======== #
# pylint: disable=W0603

"""
A header module that contains the code required to make transpiled calligraphy
scripts run
"""

import subprocess
import os
import sys
import base64
from typing import Union
import importlib.util

class Environment:
    """A class to act as a convenient method to access environment variables"""

    def __init__(self) -> None:
        """Initialize the Environment object"""

    def __getattribute__(self, name: str) -> str:
        """Retrieve an environment variable by name

        Args:
            name (str): Name of the environment variable to get

        Returns:
            str: Value of the environment variable accessed
        """

        return os.getenv(name)

    def __setattr__(self, name: str, value: str) -> None:
        """Set and environment variable to the given value

        Args:
            name (str): Name of the environment variable to set
            value (str): Value to set the environment variable to
        """

        os.environ[name] = value


class Options:
    """A class to make setting of bash shell options more convenient"""

    def __init__(self) -> None:
        """Initialize the Options object and set all shell options to default values"""

        # ignores:
        # * -n
        # * -o noexec
        # * -o emacs
        # * -o vi
        self.a = False
        self.b = False
        self.e = True  # non-default
        self.f = False
        self.h = True
        self.k = False
        self.m = False
        self.p = False
        self.t = False
        self.u = False
        self.v = False
        self.x = False
        self.B = True
        self.C = False
        self.E = False
        self.H = True
        self.P = False
        self.T = False
        self.history = True
        self.ignoreeof = False
        self.pipefail = True  # non-default
        self.posix = False

        self.keys = [
            "a",
            "b",
            "e",
            "f",
            "h",
            "k",
            "m",
            "p",
            "t",
            "u",
            "v",
            "x",
            "B",
            "C",
            "E",
            "H",
            "P",
            "T",
            "history",
            "ignoreeof",
            "pipefail",
            "posix",
        ]

    def bash_string(self) -> None:
        """Construct the set command for all the options and their values"""

        true_single_opts = [
            key for key in self.keys if getattr(self, key) == True and len(key) == 1
        ]
        false_single_opts = [
            key for key in self.keys if getattr(self, key) == False and len(key) == 1
        ]
        true_multiple_opts = [
            key for key in self.keys if getattr(self, key) == True and len(key) > 1
        ]
        false_multiple_opts = [
            key for key in self.keys if getattr(self, key) == False and len(key) > 1
        ]

        string = "set "
        if len(true_single_opts) > 0:
            string += "-"
            for opt in true_single_opts:
                string += opt
            string += " "
        if len(false_single_opts) > 0:
            string += "+"
            for opt in false_single_opts:
                string += opt
            string += " "
        for opt in true_multiple_opts:
            string += f"-o {opt} "
        for opt in false_multiple_opts:
            string += f"+o {opt} "

        return string


def source_import(calligraphy_path, module_name):
    if not calligraphy_path.endswith(".script"):
        raise ImportError(
            "Calligraphy only support sourcing other scripts that end with the '.script' extension"
        )
    if not os.path.exists(calligraphy_path):
        raise FileNotFoundError(
            f"Sourced script of '{calligraphy_path}' does not exist"
        )

    directory, script = os.path.split(calligraphy_path)
    python_path = os.path.join(directory, f".{script[:-7]}.py")

    spec = importlib.util.spec_from_file_location(module_name, python_path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[module_name] = module
    spec.loader.exec_module(module)
    return module


RC = 0
env = Environment()
shellopts = Options()

def shell(
    cmd: str,
    get_rc: bool = False,
    get_stdout: bool = False,
    silent: bool = False,
    format_dict: dict = {},
) -> Union[None, str, int]:
    """Perform a shell call and update the environment with any env variable changes

    Args:
        cmd (str): The command to run
        get_rc (bool, optional): Should the return code of the call be returned.
            Defaults to False.
        get_stdout (bool, optional): Should the contents of stdout of the call be
            returned. Defaults to False.
        silent (bool, optional): Should the output to stdout be suppressed when printing
            to the terminal. Defaults to False.
        format_dict (dict): Dictionary of values to use in command formatting

    Raises:
        RuntimeError: The shell command exited with a non-zero return code when not in
            an if statement where the RC is being explicitly checked

    Returns:
        Union[None, str, int]: Default None, stdout contents if get_stdout is True and
            return code if get_rc is True
    """

    global RC
    global env

    cmd_bytes = cmd.encode("utf-8")
    decoded_bytes = base64.b64decode(cmd_bytes)

    decoded = decoded_bytes.decode("utf-8")
    decoded = decoded.format(**format_dict)

    decoded = f"{shellopts.bash_string()} && {decoded} && echo '\n' && echo ~~~~START_ENVIRONMENT_HERE~~~~ && printenv && echo ~~~~START_CWD_HERE~~~~ && pwd"

    decoded_bytes = decoded.encode("utf-8")
    cmd_bytes = base64.b64encode(decoded_bytes)
    cmd = cmd_bytes.decode("utf-8")

    cmd = f"echo '{cmd}' | base64 -d | bash"
    stdout = []
    envout = []
    cwd_path = os.getcwd()

    with subprocess.Popen(
        cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, env=os.environ.copy()
    ) as proc:
        # grab and return the exit code
        is_stdout = True
        is_env = False
        for line in iter(proc.stdout.readline, b""):
            str_line = line.decode("utf-8")[:-1]
            if str_line == "~~~~START_ENVIRONMENT_HERE~~~~":
                if len(stdout) > 1:
                    if stdout[-2]:
                        print(stdout[-2])
                    stdout = stdout[:-1]
                is_stdout = False
                is_env = True
            elif str_line == "~~~~START_CWD_HERE~~~~":
                is_env = False
            elif is_stdout:
                if not silent and len(stdout) > 1:
                    print(stdout[-2])
                stdout.append(str_line)
            elif is_env:
                envout.append(str_line)
            else:
                cwd_path = str_line
        proc.stdout.close()
        proc.wait()
        RC = proc.poll()

    env.CALLIGRAPHY_RC = str(RC)

    # change our directory to where the shell command took us
    os.chdir(cwd_path)

    # update environment with what was modified by the shell command
    for line in envout:
        line = line.strip().split("=")
        if len(line) > 1:
            os.environ[line[0]] = line[1]

    # we don't want to raise exceptions if a user is checking for the return code
    # explicitly
    if not get_rc and shellopts.e and RC != 0:
        raise RuntimeError(f"The shell command failed with return code {RC}")

    if get_stdout:
        return "\n".join(stdout)
    if get_rc:
        return RC
    return None

# ======== END HEADER FILE ======== #

def run_command(command: str) -> str:
    out = shell(base64.b64encode(command.encode('utf-8')).decode('utf-8'), get_stdout=True, silent=True)
    return out

def get_pods(config: dict=None) -> dict[str, str]:
    pods = run_command(f'kubectl --namespace {config["namespace"]} --context {config["context"]} get pods -o json')
    pod_data = json.loads(pods)

    out = {}
    for item in pod_data['items']:
        name = item['metadata']['name']
        for pattern in config['pod_regex']:
            if re.match(pattern, name):
                out[name] = item['status']['phase']
    
    return out

def check_random(a: list[str] = None, b: list[str] = None, jump_max: int = 1) -> tuple[bool, list[str]]:
    idx = 0
    iteration = 0
    while idx < len(a) and idx < len(b):
        if not a[idx] == b[idx]:
            return False, None
        if iteration < 10:
            idx += 1
        else:
            idx += random.randint(1, jump_max)
        iteration += 1
    # a is our log window and as such it is strictly of a lesser length than b
    return True, b[len(a):]

def get_latest_logs(log_window: list=None, logs: list[str]=None) -> tuple[list[str], list[str]]:
    if not log_window:
        return logs[:], logs[:]

    window_offset = 0
    while window_offset < len(log_window):
        match, new_logs = check_random(log_window[window_offset:], logs)
        if match:
            return logs[:], new_logs[:]
        window_offset += 1
    
    return logs[:], logs[:]

def log_alert(log: str, alerts: dict[str, str], group: str="default") -> None:
    for pattern in alerts:
        if re.match(pattern, log):
            print(f'error::{group}::log_alert::{alerts[pattern]}::{log}')

def pod_alert(pods: dict[str, str]=None, patterns: dict[str, int]=None, group: str="default") -> None:
    for pattern in patterns:
        running_count = 0
        required_count = patterns[pattern]
        for pod in pods:
            status = pods[pod]
            if status == 'Running':
                running_count += 1
                continue
            if not status in ['ContainerCreating', 'Terminating']:
                print(f'error::{group}::pod_status::{pod}::{status}')
        if running_count < required_count:
            print(f'error::{group}::pod_count::{pod}::{running_count}<{required_count}')

if __name__ == '__main__':
    '''
    config argument should be of format:
        {
            "namespace": str, // k8s namespace to use
            "pod_regex": {
                "pattern": int // this is the number of pods that should be in a running status
            },
            "context": str, // k8s context to use
            "window_size": int, // how many logs to hold in the window for new log detection
            "interval": int, // how many seconds should the monitor wait before polling again
            "group": str, // what group to put in error alerts, e.g. `error::<your group>::pod_status::<pod name>`...
            "alert_regex": [
                {
                    "pattern": str, // log line pattern to set off the alert
                    "alert": str // alert type, e.g. error::foo::bar
                }
            ]
        }
    '''

    # Used for shell execution
    sys.argv = "PROGRAM_ARGS"

    # Get the pods in this namespace/context that will be monitored
    pods = get_pods(config)

    windows = {}
    for pod in pods:
        windows[pod] = []

    while True:
        pods = get_pods(config)
        pod_alert(pods, config['pod_regex'], group=config['group'])
        for pod in pods:
            if not pod in windows:
                windows[pod] = []
        to_remove = []
        for pod in windows:
            if not pod in pods:
                to_remove.append(pod)
                logs = [l for l in run_command(f'kubectl --namespace {config["namespace"]} --context {config["context"]}  logs --previous --tail={config["window_size"]} {pod}').split('\n') if l]
            else:
                logs = [l for l in run_command(f'kubectl --namespace {config["namespace"]} --context {config["context"]}  logs --tail={config["window_size"]} {pod}').split('\n') if l]
            windows[pod], new_logs = get_latest_logs(windows[pod], logs)
            if not new_logs:
                new_logs = []
            for log in new_logs:
                # print(log)
                log_alert(log, config['alert_regex'], config['group'])
            
        for pod in to_remove:
            del(windows[pod])
        time.sleep(config["interval"])
