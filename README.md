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

`bro` allows you to execute load testing scenarios with both constant and variable rate traffic patterns.

During the execution it collects metrics (e.g. RPS, latency, errors) and validates them against defined thresholds. 

It is written in [Go](https://github.com/golang/go), test scenarios are defined in YAML.

Try it together with [mox](https://github.com/lameaux/mox) - a tool to stub external dependencies, so that your
application can be tested in isolation.

Check out [nft](https://github.com/lameaux/nft) repo to learn more about **bro** & **mox** for non-functional testing.

# Installation

Make sure you have [Go](https://go.dev/doc/install) installed and `GOPATH` is set up correctly.

Clone this repository and run:

```shell
make install
```

You will have `bro`, `brod` and `broctl` installed.

# Usage

## bro

`bro` is the main binary that runs load testing scenarios. 

Optionally, it can be integrated with `brod` to aggregate statistics and run distributed tests with `broctl`. 

```shell
bro [flags] <config.yaml>

--debug
--silent
--logJson
--skipBanner
--skipResults
--skipExitCode
--brodAddr=brod:8080
```

### Flags

#### --debug

Enables debug mode. Results in a more detailed logging.

#### --silent

Changes log level from INFO to ERROR.

#### --logJson

Changes log format to JSON.

#### --skipBanner

Skips printing banner to stdout.

#### --skipResults

Skips printing results table to stdout.

#### --skipExitCode

Do not return exit code 1 when tests fail.

#### --brodAddr=brod:8080

Connects `bro` (client) with `brod` (server).

### Example

```shell
bro --skipBanner --debug examples/00-ping-google.yaml

{"level":"info","version":"v0.0.1","build":"76d613c","time":"2024-08-30T19:58:05+03:00","message":"bro"}
{"level":"info","configName":"Ping Google","configFile":"examples/00-ping-google.yaml","time":"2024-08-30T19:58:05+03:00","message":"config loaded"}
{"level":"debug","disableKeepAlive":false,"timeout":5000,"maxIdleConnsPerHost":100,"time":"2024-08-30T19:58:05+03:00","message":"creating http client"}
{"level":"info","execution":"","time":"2024-08-30T19:58:05+03:00","message":"executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate."}
{"level":"info","scenario":{"name":"Check 301 Redirect","rps":1,"threads":1,"queueSize":1,"duration":1000},"time":"2024-08-30T19:58:05+03:00","message":"running scenario"}
{"level":"debug","threadId":0,"msgId":1,"method":"","url":"https://google.com","code":301,"latency":867,"checks":[{"type":"httpCode","name":"","value":"301","ok":true},{"type":"httpHeader","name":"Location","value":"https://www.google.com/","ok":true}],"success":true,"time":"2024-08-30T19:58:06+03:00","message":"response"}
{"level":"debug","threadId":0,"time":"2024-08-30T19:58:06+03:00","message":"shutting down"}
{"level":"info","scenario":{"name":"Check 200 OK","rps":1,"threads":1,"queueSize":1,"duration":1000},"time":"2024-08-30T19:58:06+03:00","message":"running scenario"}
{"level":"debug","threadId":0,"msgId":1,"method":"","url":"https://www.google.com","code":200,"latency":800,"checks":[{"type":"httpCode","name":"","value":"200","ok":true},{"type":"httpBody","name":"","value":"<!doctype html><html itemscope=\"\" itemtype=\"http://schema.org/WebPage\" lang=\"en-CY\"><head><meta cont...","ok":true}],"success":true,"time":"2024-08-30T19:58:07+03:00","message":"response"}
{"level":"debug","threadId":0,"time":"2024-08-30T19:58:07+03:00","message":"shutting down"}
{"level":"info","scenario":{"name":"Check Error","rps":1,"threads":1,"queueSize":1,"duration":1000},"time":"2024-08-30T19:58:07+03:00","message":"running scenario"}
{"level":"debug","threadId":0,"msgId":1,"method":"","url":"https://www.google.com/unknown","code":404,"latency":150,"checks":[{"type":"httpCode","name":"","value":"404","ok":true},{"type":"httpBody","name":"","value":"<!DOCTYPE html>\n<html lang=en>\n  <meta charset=utf-8>\n  <meta name=viewport content=\"initial-scale=1...","ok":true}],"success":true,"time":"2024-08-30T19:58:07+03:00","message":"response"}
{"level":"debug","threadId":0,"time":"2024-08-30T19:58:08+03:00","message":"shutting down"}
{"level":"info","totalDuration":3002.86425,"ok":true,"time":"2024-08-30T19:58:08+03:00","message":"results"}
Ping Google
┌────────────────────┬───────┬──────┬─────────┬────────┬─────────┬─────────┬──────────────┬──────────────┬─────┬────────┐
│ SCENARIO           │ TOTAL │ SENT │ SUCCESS │ FAILED │ TIMEOUT │ INVALID │ LATENCY @P99 │     DURATION │ RPS │ PASSED │
├────────────────────┼───────┼──────┼─────────┼────────┼─────────┼─────────┼──────────────┼──────────────┼─────┼────────┤
│ Check 301 Redirect │     1 │    1 │       1 │      0 │       0 │       0 │ 867 ms       │ 1.001114208s │   1 │ OK     │
│ Check 200 OK       │     1 │    1 │       1 │      0 │       0 │       0 │ 800 ms       │ 1.000531875s │   1 │ OK     │
│ Check Error        │     1 │    1 │       1 │      0 │       0 │       0 │ 150 ms       │ 1.001080875s │   1 │ OK     │
└────────────────────┴───────┴──────┴─────────┴────────┴─────────┴─────────┴──────────────┴──────────────┴─────┴────────┘
Total duration: 3.00286425s
OK
```

## brod

`brod` is a server that collects statistics from `bro` client instances and exposes Prometheus metrics. 

It is also used by `broctl` to synchronize `bro` instances when running distributed tests.

```shell
brod [flags]

--debug
--logJson
--skipBanner
--port=8080
--metricsPort=9090
```

### Flags

#### --debug

Enables debug mode. Results in more detailed logging.

#### --logJson

Changes log format to JSON.

#### --skipBanner

Skips printing banner to stdout.

#### --port=8080

Defines a port for grpc server.

#### --metricsPort=9090

Defines a port for metrics endpoint.

### Example

```shell
brod --logJson
```

## broctl

`broctl` is used to run distributed tests using several instances of `bro`. 

```shell
broctl [flags] <command>

--debug
--logJson
--skipBanner
```

### Flags

#### --debug

Enables debug mode. Results in more detailed logging.

#### --logJson

Changes log format to JSON.

#### --skipBanner

Skips printing banner to stdout.

### Example

```shell
broctl --debug
```

# Examples

See [examples](./examples) dir for testing examples:

- [Ping Google](./examples/00-ping-google.yaml)
- [Simple Constant Rate Example](./examples/01-simple-constant-rate.yaml)

Check out [nft repo](https://github.com/lameaux/nft) for more examples.

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
    rps: 50 # int
    duration: 15s # duration
    threads: 20 # int
    queue: 200 # int
    payloadType: http # only http for now
    httpRequest:
      url: http://0.0.0.0:8080/random # url
      method: GET # GET, POST, HEAD, DELETE, etc.
    checks:
      - type: httpCode
        equals: 200 # int
    thresholds:
      - name: check
        type: httpCode
        minRate: 1.0 # float
```


