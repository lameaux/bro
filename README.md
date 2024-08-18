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

```shell
bro --skipBanner --debug examples/00-ping-google.yaml

{"level":"info","version":"v0.0.1","build":"967d059","time":"2024-08-18T14:18:54+03:00","message":"bro"}
{"level":"info","configName":"Ping Google","configFile":"examples/00-ping-google.yaml","time":"2024-08-18T14:18:54+03:00","message":"config loaded"}
{"level":"debug","disableKeepAlive":false,"timeout":5000,"maxIdleConnsPerHost":100,"time":"2024-08-18T14:18:54+03:00","message":"creating http client"}
{"level":"info","execution":"","time":"2024-08-18T14:18:54+03:00","message":"executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate."}
{"level":"info","scenario":{"name":"Check Redirect","rate":1,"interval":1000,"vus":1,"duration":1000},"time":"2024-08-18T14:18:54+03:00","message":"running scenario"}
{"level":"debug","vuId":0,"msgId":1,"method":"GET","url":"http://google.com","code":301,"latency":156,"expectedCode":301,"success":true,"time":"2024-08-18T14:18:55+03:00","message":"response"}
{"level":"debug","vuId":0,"time":"2024-08-18T14:18:55+03:00","message":"shutting down"}
{"level":"info","totalDuration":1000.919667,"time":"2024-08-18T14:18:55+03:00","message":"results"}

Total duration: 1.000919667s
┌────────────────┬────────────────┬──────┬────────────┬────────┬─────────┬─────────┬──────────────┬─────────────┐
│ SCENARIO       │ TOTAL REQUESTS │ SENT │ SUCCESSFUL │ FAILED │ TIMEOUT │ INVALID │ LATENCY @P99 │ DURATION    │
├────────────────┼────────────────┼──────┼────────────┼────────┼─────────┼─────────┼──────────────┼─────────────┤
│ Check Redirect │              1 │    1 │          1 │      0 │       0 │       0 │ 156 ms       │ 1.00090475s │
└────────────────┴────────────────┴──────┴────────────┴────────┴─────────┴─────────┴──────────────┴─────────────┘
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


