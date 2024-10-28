# bro

**bro** is a load testing tool.

![Screenshot](.github/images/bro1.png)

## Overview

**bro** enables you to run load testing scenarios with both constant rate and ramp-up traffic patterns.
During execution, it collects metrics such as requests per second (RPS), latency, and errors, and validates them against defined thresholds. 

It is written in [Go](https://github.com/golang/go), test scenarios are defined in YAML.

Try it together with **[mox](https://github.com/lameaux/mox)**, a tool for stubbing external dependencies, to test your application in isolation.

## Installation

Make sure you have [Go](https://go.dev/doc/install) installed and `GOPATH` is set up correctly.

Clone this repository and run `make install`:

```shell
git clone https://github.com/lameaux/bro.git
cd bro
make install
```

You will have `bro`, `brod` and `broctl` installed.

See [User Manual](docs/user-manual.md) for instructions on how to use these tools.

## Examples

See [Examples](./examples/README.md) for basic scenarios. 
More advanced scenarios can be found in [NFT repo](https://github.com/lameaux/nft). 

Check it out to learn more about using **bro** & **mox** for non-functional testing.

![Screenshot](.github/images/bro2.png)



