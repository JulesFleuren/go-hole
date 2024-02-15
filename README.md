# go-hole
[![Build Status](https://travis-ci.org/davidepedranz/go-hole.svg?branch=master)](https://travis-ci.org/davidepedranz/go-hole)
[![codecov](https://codecov.io/gh/davidepedranz/go-hole/branch/master/graph/badge.svg)](https://codecov.io/gh/davidepedranz/go-hole)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidepedranz/go-hole)](https://goreportcard.com/report/github.com/davidepedranz/go-hole)

`go-hole` is a fast and lightweight [DNS sinkhole](https://en.wikipedia.org/wiki/DNS_Sinkhole) that blocks domains known to serve ads, tracking scripts, malware and other unwanted content. It also caches DNS responses to reduce latency, and collects anonymous statistics about the DNS traffic. `go-hole` is written in Go and runs on every platform and operating systems supported by the Go compiler. `go-hole` can be combined with a private VPN to protect mobile devices on every network.

## TL;TR

Run as a Docker container and use as your primary DNS server:
```sh
docker run --name go-hole -d -p 127.0.0.1:53:8053/udp davidepedranz/go-hole:latest
```

Test that `go-hole` is working correctly:
```sh
nslookup -port=8053 example.com localhost
nslookup -port=8053 googleadservices.com localhost
```

## How does it work?

`go-hole` runs a custom DNS server that selectively blocks unwanted domains by replying `NXDomain (Non-Existent Domain)` to the client. It uses an upstream DNS (by default DNS over Tls via cloudflare [1.1.1.1](https://1.1.1.1/)) to resolve the queries the first time, then it caches the response to speed up the following requests.

## Why?

The amount of intrusive ads and tracking services on the Internet is huge and continues to grow. While it is quite easy to block them on a computer using your favourite ad-block plugin, it is difficult or even impossible to do the same on mobile devices. This project aims to block unwanted ads and services at the network level, without the need to install any software on the user's device.

This project is inspired by [Pi-Hole](https://github.com/pi-hole/pi-hole).

## Build & Run

```sh
# build the binary
go build

# execute the binary
# please make sure the blacklist is available at ./data/blacklist.txt
./go-hole
```

## Configuration

`go-hole` can be configured using a few environment variables:

| Environment Variable   | Default Value | Function                                                              |
| ---------------------- | ------------- | --------------------------------------------------------------------- |
| `DNS_PORT`             | `53`          | UDP port where to listen for DNS queries.                             |
| `PROMETHEUS_PORT`      | `9090`        | TCP port where to serve the collected metrics. Port 0 disables the service.                        |
| `UPSTREAM_DNS`         | `1.1.1.1:853`  | IP and port of the upstream DNS to use to resolve the queries.        |
| `UPSTREAM_TLS_SRVNAME` | `one.one.one.one`            | DNS server name for TLS certificate validation (enables DNS over TLS)  |
| `DEBUG`                | `false`       | If true, `go-hole` logs all queries to the standard output.           |

By default no domains are blacklisted. You can import blacklists from online hosted sources, such as those listed on [firebog.net](https://firebog.net/). To import such a list visit the server on port 8080 and paste the url's in the text box. Once you press `Save changes`, the dns server will be restarted and will now block the domains from the sources.

## Web interface

Some settings of go-hole can be changed through the webinterface which is accesible on port 8080. The interface is password protected, but it is an http connection, so someone who knows what they're doing could steal your login details. So **ONLY USE THIS ON A TRUSTED HOME NETWORK AND DON'T EXPOSE IT TO THE INTERNET.** The default username and password are both `admin`, see this [FAQ](#change-password) on how to change this.

## FAQ

### Do you have a Docker container?

Sure, checkout the automatic build on Docker Hub: https://hub.docker.com/r/davidepedranz/go-hole/

### Can I combine it with a VPN software?

Sure, this is the main setup of `go-hole`. For example, you can combine it with [OpenVPN](https://openvpn.net/). We will publish soon a guide to setup `go-hole` and OpenVPN together on a private server.

### Change Password

Changing the username and password can be done either via the environmental variables, or via the configuration file. You need to set the values of `ADMIN_USR_HASH` and `ADMIN_PWD_HASH` to a bcrypt hash of the desired username and password respectively. Once again a warning that the credentials are sent over http, so don't use a password that you use for anything else. The hashes can be generated with the following code snippet. If you set the cost higher, the hash will be harder to crack, but it increases the time it takes to verify the password. Since the hash is most likely not the weakest link in this system, there is not a lot of use in setting it higher.


```
package main

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cost := 4
	hash, _ := bcrypt.GenerateFromPassword([]byte("<username/passwd>"), cost)
	fmt.Println(base64.StdEncoding.EncodeToString(hash[:]))
}
```
Tip: if you don't have go installed you can use https://go.dev/play/.

### Privacy Issues

By default, `go-hole` does not log any DNS query. Logging can be enabled for debug purposes, but we discourage it in production, since it breaches the privacy of the users. On the other hand, `go-hole` is fully instrumented to collect anonymous data about the amount of blocked queries, the response times and other performance metrics.

### Metrics

`go-hole` is instrumented with [Prometheus](https://prometheus.io/) to collect the following metrics:

| Type      | Name                                       | Help                                          |
| --------- | ------------------------------------------ | --------------------------------------------- |
| Histogram | `gohole_dns_queries_duration_seconds`      | Duration of replies to DNS queries.           |
| Histogram | `gohole_blacklist_lookup_duration_seconds` | Duration of a domain lookup in the blacklist. |
| Histogram | `gohole_cache_operation_duration_seconds`  | Duration of an operation on the cache.        |
| Histogram | `gohole_override_duration_seconds`         | Duration of a domain overrided lookup.        |

By default, metrics are served over HTTP at port `9090` and path `/metrics`.

## License

`go-hole` is free software released under the MIT Licence. Please checkout the [LICENSE](./LICENSE) file for details.
