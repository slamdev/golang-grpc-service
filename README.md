# golang-http-service
Template for creating GRPC service

## Components

### Configuration

Application configuration is defined [application.yaml](configs/application.yaml) file. There is **profiles** system that
allows to use and merge multiple configuration files together. It is controlled with **ACTIVE_PROFILES** environment
variable. E.g.: `ACTIVE_PROFILES=cloud,dev` will merge these files together:
1. application.yaml
2. application-cloud.yaml
3. application-dev.yaml

[Uber config](https://github.com/uber-go/config) library is used to parse and merge config files. It also supports 
environment variables.

Configuration files are embedded into the resulting binary with [pkger](https://github.com/markbates/pkger) tool.

### Monitoring

[Zap](https://github.com/uber-go/zap) is used to control logs. Logs are outputted in plain text format when the
application is running locally and in json format when the **cloud** profile is used.

Application exposes [Prometheus](https://prometheus.io/) metrics at **/metrics** endpoint.

Application exposes health endpoint at **/health**.

### Testing

For reach assertions the [testify](https://github.com/stretchr/testify) library is used.

### Linting

[Spectral](https://github.com/stoplightio/spectral) is used to lint OpenAPI files and 
[golangci-lint](https://github.com/golangci/golangci-lint) for Go files.

### Building

All the build process is describe it the (Makefile)[Makefile]. Run `make build` to test and build the binary.
