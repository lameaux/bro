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

Check out [nft](https://github.com/lameaux/nft) repo for non-functional test examples.

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

{"level":"info","version":"v0.0.1","build":"e2f4baa","time":"2024-08-18T20:32:46+03:00","message":"bro"}
{"level":"info","configName":"Ping Google","configFile":"examples/00-ping-google.yaml","time":"2024-08-18T20:32:46+03:00","message":"config loaded"}
{"level":"debug","disableKeepAlive":false,"timeout":5000,"maxIdleConnsPerHost":100,"time":"2024-08-18T20:32:46+03:00","message":"creating http client"}
{"level":"info","execution":"","time":"2024-08-18T20:32:46+03:00","message":"executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate."}
{"level":"info","scenario":{"name":"Check 301 Redirect","rate":1,"interval":1000,"vus":1,"duration":1000},"time":"2024-08-18T20:32:46+03:00","message":"running scenario"}
{"level":"debug","vuId":0,"msgId":1,"method":"GET","url":"https://google.com","code":301,"latency":249,"expectedCode":301,"success":true,"time":"2024-08-18T20:32:46+03:00","message":"response"}
{"level":"debug","vuId":0,"time":"2024-08-18T20:32:47+03:00","message":"shutting down"}
{"level":"info","scenario":{"name":"Check 200 OK","rate":1,"interval":1000,"vus":1,"duration":1000},"time":"2024-08-18T20:32:47+03:00","message":"running scenario"}
{"level":"debug","vuId":0,"msgId":1,"method":"GET","url":"https://www.google.com","code":200,"latency":269,"expectedCode":200,"success":true,"time":"2024-08-18T20:32:47+03:00","message":"response"}
{"level":"debug","vuId":0,"time":"2024-08-18T20:32:48+03:00","message":"shutting down"}
{"level":"info","scenario":{"name":"Check Error","rate":1,"interval":1000,"vus":1,"duration":1000},"time":"2024-08-18T20:32:48+03:00","message":"running scenario"}
{"level":"debug","vuId":0,"msgId":1,"method":"GET","url":"https://www.google.com/unknown","code":404,"latency":148,"expectedCode":200,"success":false,"time":"2024-08-18T20:32:48+03:00","message":"response"}
{"level":"debug","vuId":0,"time":"2024-08-18T20:32:49+03:00","message":"shutting down"}
{"level":"info","totalDuration":3002.783667,"ok":true,"time":"2024-08-18T20:32:49+03:00","message":"results"}
┌────────────────────┬───────┬──────┬─────────┬────────┬─────────┬─────────┬──────────────┬──────────────┬─────┬────────┐
│ SCENARIO           │ TOTAL │ SENT │ SUCCESS │ FAILED │ TIMEOUT │ INVALID │ LATENCY @P99 │     DURATION │ RPS │ PASSED │
├────────────────────┼───────┼──────┼─────────┼────────┼─────────┼─────────┼──────────────┼──────────────┼─────┼────────┤
│ Check 301 Redirect │     1 │    1 │       1 │      0 │       0 │       0 │ 249 ms       │ 1.001176917s │   1 │ OK     │
│ Check 200 OK       │     1 │    1 │       1 │      0 │       0 │       0 │ 269 ms       │ 1.000128625s │   1 │ OK     │
│ Check Error        │     1 │    1 │       0 │      1 │       0 │       1 │ 148 ms       │ 1.001192375s │   1 │ OK     │
└────────────────────┴───────┴──────┴─────────┴────────┴─────────┴─────────┴──────────────┴──────────────┴─────┴────────┘
Total duration: 3.002783667s
OK
```

# Examples

See [examples](./examples) dir for testing examples:

- [Ping Google](./examples/00-ping-google.yaml)
- [Simple Constant Rate Example](./examples/01-simple-constant-rate.yaml)

Check out [nft](https://github.com/lameaux/nft) repo for more examples.

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
    validate:
      failed:
        eq: 0 # int
```


