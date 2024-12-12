# Task

Tasks are the actual execution blocks of the workflows. These tasks can either be run directly on the node (helpful in k8s when it doesn't want to play nice with container-in-container) or by spawning an internal container.

Additionally, tasks depend on others in order to form a Directed Acyclic Graph (DAG). If tasks are set to auto-execute, they then will trigger on all their parent tasks meeting their success criteria.
For example, if we depend on `A` succeeding and `B` failing, then our task will only run when both those parent tasks have met those conditions

Finally, tasks work off a filestore and context. The filestore is where you can upload scripts and such that you may want to use in your tasks. Via your task configuration you can pull down files to use or store others for future use

The context can be considered as a dictionary of key-value pairs which you can load into your task for use as well as persist to, which will then be passed down to child tasks. This makes it useful for storing a calculated value for use by later steps

## Schema

```yaml
name: str # task name
kind: str # kind of task to execute. `local|container`
container_login_command: str # login command to execute before pulling container, e.g. login to DockerHub
auto_execute: bool # should the task execute on all depends_on matching their success conditions. defaults to `false`
should_rm: bool # should the task remove the execution container after finishing. defaults to `false`. only used with `container` kind
image: str # container image to run task in, only used with `container` kind
disabled: bool # is the task disabled from execution. defaults to `false`
depends_on: # [optional] tasks to depend on execution status for auto-trigger/layout
  success:
    - str # task name to depend on success status
  error:
    - str # task name to depend on error status
  always:
    - str # task name to depend on success or error status
run: | # task code to execute
  str
env: # [optional] ENV vars to include in the task execution
  str: str # ENV VAR NAME: value
store: # [optional] persist assets from an executed task
  env:
    - str # name of value to persist in context (name is ENV variable name)
  file:
    - str # name of file to load from filestore
load: # [optional] include assets for task execution
  env:
    - str # values to load in from the context (name is context variable name)
  file:
    - str # name of file to store to filestore
inputs: # input values to load into the task
  str: str # ENV VAR NAME: Input name
```
