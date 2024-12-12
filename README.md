# Scaffold

## About

Scaffold is an infrastructure management tool that takes a waterfall approach to management. This works by tracking the state of each workflow (waterfall workflow DAG) task and triggering the next task as all its parents are in a success state. Workflows are manually triggered and report their state to the UI so engineers can track deployment and upgrade processes to ensure that they work as expected. Additionally, Scaffold allows for input and file storage and loading to make task execution easier. Finally, tasks are executed within container images to enable dependencies to be brought along to the task execution.

## Rationale

Infrastructure management is an interesting problem to tackle. Many organization use CI/CD pipelines to handle deployment and upgrades (e.g. Jenkins, Concourse) which allows for an automated approach to management, however execution status and state can be hard to track. Additionally, changes that may be inside the middle of an upgrade pipeline may need to be re-run without running the whole pipeline and may change the required state of subsequent tasks. The waterfall approach allows for independent execution of tasks with their state changes propagating down to dependent tasks.

## Getting Started

See the [docs](https://scaffold.readthedocs.io/en/latest) for information on getting started and reference

## Issues

If you have any questions or concerns please reach out at scaffoldworkflow@gmail.com or create an [issue](https://github.com/scaffoldworkflow/scaffold/issues)

## License

Scaffold is licensed under the MIT license
