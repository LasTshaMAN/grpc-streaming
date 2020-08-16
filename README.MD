# GRPC-streaming

Table of contents:

- [Architecture](docs/architecture.md)
TODO
- [Dependencies](#dependencies)
- [How to run](#how-to-run)
- [Things yet to be done](#things-yet-to-be-done-)

This project is an experimental ground for GRPC streaming.

## Dependencies

- `Golang` 1.14 or higher
- `protoc` 3.0.0 or higher
- `grpc plugin` for `protoc` (latest `master` version from [grpc-go](https://github.com/grpc/grpc-go) repo, see explanation below)
- `docker-compose`

To start developing with GRPC take a look at [quickstart guide](https://grpc.io/docs/languages/go/quickstart/).
We are using `master` version of [grpc-go](https://github.com/grpc/grpc-go) repo due to some temporary issues,  
see a note from [quickstart guide](https://grpc.io/docs/languages/go/quickstart/):
```
We are in the process of transitioning to a new Go protoc plugin. Until the transition is complete, you need to install grpc-go/cmd/protoc-gen-go-grpc manually before regenerating .pb.go files. To track progress on this issue, see Update Go quick start #298.
```
and [this comment](https://github.com/grpc/grpc.io/issues/298#issuecomment-656767393).

### How to run

To run linter use:
```
make lint_docker
```

To run unit-tests do:
```
# Spin up service dependencies (for example, Redis) first:
make up_deps

make test
```

To deploy [server](./cmd/server) locally with `docker-compose` do:
```
# Spin up service dependencies (for example, Redis) first:
make up_deps

make run_server
```

You can also run a [client](./cmd/client) to send some requests to the server (check out [logs](./logs) to see what happens):
```
make run_client
```

See [Makefile](./Makefile) for the full list of available commands.

### Things yet to be done ...
- fix linter suggestions
- overall architecture documentation
- metrics
- extend unit-test coverage (especially for edge and error cases)
- tracing
- move all the settings (such as timeouts, amount of connections, locks, ...) in the configuration file
- customize timeouts according to our needs
- setup CI/CD
