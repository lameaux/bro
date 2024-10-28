# Test Configuration

```yaml
name: Example Config # string
parallel: false # only serial for now
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


