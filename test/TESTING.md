# Testing

Requires that `gocovmerge` is available, to install run:

```shell
go install github.com/wadey/gocovmerge@latest
```

To run integration tests:

```bash
stud build-docker-test && stud run-test-both
```
