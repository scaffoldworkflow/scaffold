# Workflow

Scaffold workflows are the core of the tool's functionality. These workflows allow for you to define a collection of tasks and their execution dependencies with manual inputs to be used to control manually triggered runs.

## Schema

```yaml
version: str # workflow api version
name: str # workflow name
groups:  # workflow groups that can view this workflow
  - str
inputs:
  - ...
tasks:
  - ...
```
