# Scaffold

## About

// TODO: Fill this out

## Rationale

// TODO: Fill this out

## Getting Started

See the [docs](https://scaffold.readthedocs.io/en/latest) for information on getting started and reference

## Issues

If you have any questions or concerns please reach out at scaffoldworkflow@gmail.com or create an [issue](https://github.com/scaffoldworkflow/scaffold/issues)

## License

Scaffold is licensed under the MIT license

# TODO

## 0.4.2

- [x] Auto prune old jobs
- [x] On start check for any runs in progress and start a monitor for them
- [ ] Single project UI page
- [x] Script path for run contents
- [x] Mount worker scripts instead of worker pod
- [ ] Update CLI to apply resources
- [ ] Update CLI to trigger release promotion
- [ ] Add promote step page that shows output and status instead of directing to the workflow page

## 0.4.3

- [ ] Telemetry functionality
- [ ] Unify endpoints with query params
- [ ] Move kernel execution to k8s job
- [ ] Integrate project resources into runbooks
- [ ] Integrate project resources into alerts
- [ ] Proper permission management
- [ ] Helm chart

## 0.4.4

- [ ] Update Python client to match new setup
- [ ] Fix tests
- [ ] Move resource type and resource definitions to be global in project?

## 0.5.0

- [ ] Use regular step definition for promote jobs
- [ ] Ability to re-run step with shell
- [ ] Auto-setup jq on images where not present
- [ ] Okta integration
