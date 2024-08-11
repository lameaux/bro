```
 █████                       
░░███                        
 ░███████  ████████   ██████ 
 ░███░░███░░███░░███ ███░░███
 ░███ ░███ ░███ ░░░ ░███ ░███
 ░███ ░███ ░███     ░███ ░███
 ████████  █████    ░░██████ 
░░░░░░░░  ░░░░░      ░░░░░░  
```

**bro** is a load testing tool.

# About

This tool allows you to execute load testing scenarios with both constant and variable rate traffic patterns.

During the execution it collects metrics (e.g. RPS, latency, errors) and validates them
against defined thresholds. 

It is written in Go, test scenarios are defined in YAML.

Try it together with [mox](https://github.com/lameaux/mox) - a tool to stub external dependencies, so that your
application can be tested in isolation.

# Installation

Make sure you have `GOPATH` set up correctly.

```shell
make install
```

# Usage

```shell
bro [flags] <config.yaml>

--debug
--silent
--skipBanner
--skipResults
--metricsPort=9090

Examples:
- bro --silent example/02-ping-google-com.yaml 
- bro --debug examples/01-simple-constant-rate.yaml
```

# Examples

See [examples](./examples) dir for testing examples:

- [Ping Google](./examples/00-ping-google.yaml)
- [Simple Constant Rate Example](./examples/01-simple-constant-rate.yaml)

# Test Configuration

```yaml
name: Example Config # string
execution: serial # only serial for now
httpClient:
  timeout: 5s # duration
  maxIdleConnsPerHost: 100 # int
  disableKeepAlive: false # bool
  disableFollowRedirects: true # bool  
scenarios: # list
  - name: Example Scenario # Constant rate demo
    rate: 50 # int
    interval: 1s # duration
    duration: 15s # duration
    vus: 20 # int
    buffer: 200 # int
    payloadType: http # only http for now
    httpRequest:
      url: http://0.0.0.0:8080/random # url
      method: get # get, post, head, delete, etc.
    httpResponse:
      code: 200 # int
    thresholds:
      latency:
        - p: 99 # int
          v: 500 # int
```


