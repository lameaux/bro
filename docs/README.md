# User Manual

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
bro --skipBanner --debug examples/ping/google.yaml

4:50PM INF bro build=91fef7e version=v0.0.1
4:50PM INF config loaded config={"name":"Ping Google","path":"examples/00-ping-google.yaml"}
4:50PM DBG creating http client disableKeepAlive=false maxIdleConnsPerHost=100 timeout=5000
4:50PM INF executing scenarios... press Ctrl+C (SIGINT) or send SIGTERM to terminate. execution=
4:50PM INF running scenario scenario={"duration":1000,"name":"Check 301 Redirect","queueSize":1,"rps":1,"threads":1}
4:50PM DBG response checks=[{"name":"","pass":true,"type":"httpCode","value":"301"},{"name":"Location","pass":true,"type":"httpHeader","value":"https://www.google.com/"}] code=301 latency=371 method=GET msgId=1 success=true threadId=0 url=https://google.com
4:50PM DBG shutting down threadId=0
4:50PM DBG threshold validation metric=checks passed=true rate=1 scenario={"name":"Check 301 Redirect"} type=httpCode
4:50PM DBG threshold validation metric=checks passed=true rate=1 scenario={"name":"Check 301 Redirect"} type=httpHeader
4:50PM INF running scenario scenario={"duration":1000,"name":"Check 200 OK","queueSize":1,"rps":2,"threads":1}
4:50PM DBG response checks=[{"name":"","pass":true,"type":"httpCode","value":"200"},{"name":"","pass":true,"type":"httpBody","value":"<!doctype html><html itemscope=\"\" itemtype=\"http://schema.org/WebPage\" lang=\"en-CY\"><head><meta cont..."}] code=200 latency=319 method=GET msgId=1 success=true threadId=0 url=https://www.google.com
4:50PM DBG response checks=[{"name":"","pass":true,"type":"httpCode","value":"200"},{"name":"","pass":true,"type":"httpBody","value":"<!doctype html><html itemscope=\"\" itemtype=\"http://schema.org/WebPage\" lang=\"en-CY\"><head><meta cont..."}] code=200 latency=118 method=GET msgId=2 success=true threadId=0 url=https://www.google.com
4:50PM DBG shutting down threadId=0
4:50PM DBG threshold validation metric=checks passed=true rate=1 scenario={"name":"Check 200 OK"} type=httpCode
4:50PM INF running scenario scenario={"duration":1000,"name":"Check Error","queueSize":1,"rps":1,"threads":1}
4:50PM DBG response checks=[{"name":"","pass":true,"type":"httpCode","value":"404"},{"name":"","pass":true,"type":"httpBody","value":"<!DOCTYPE html>\n<html lang=en>\n  <meta charset=utf-8>\n  <meta name=viewport content=\"initial-scale=1..."}] code=404 latency=147 method=GET msgId=1 success=true threadId=0 url=https://www.google.com/unknown
4:50PM DBG shutting down threadId=0
4:50PM DBG threshold validation metric=checks passed=true rate=1 scenario={"name":"Check Error"} type=httpCode
4:50PM INF result success=true totalDuration=3003.865
Name: Ping Google
Path: examples/ping/google.yaml
┌────────────────────┬───────┬──────┬─────────┬────────┬─────────┬─────────┬──────────────┬──────────────┬─────┬────────┐
│ SCENARIO           │ TOTAL │ SENT │ SUCCESS │ FAILED │ TIMEOUT │ INVALID │ LATENCY @P99 │     DURATION │ RPS │ PASSED │
├────────────────────┼───────┼──────┼─────────┼────────┼─────────┼─────────┼──────────────┼──────────────┼─────┼────────┤
│ Check 301 Redirect │     1 │    1 │       1 │      0 │       0 │       0 │ 371 ms       │ 1.001213375s │   1 │ true   │
│ Check 200 OK       │     2 │    2 │       2 │      0 │       0 │       0 │ 319 ms       │ 1.001165167s │   2 │ true   │
│ Check Error        │     1 │    1 │       1 │      0 │       0 │       0 │ 147 ms       │  1.00031725s │   1 │ true   │
└────────────────────┴───────┴──────┴─────────┴────────┴─────────┴─────────┴──────────────┴──────────────┴─────┴────────┘
Total duration: 3.003865s
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