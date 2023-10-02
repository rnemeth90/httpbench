# httpbench [![build-release-binary](https://github.com/rnemeth90/httpbench/actions/workflows/build.yaml/badge.svg)](https://github.com/rnemeth90/httpbench/actions/workflows/build.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/rnemeth90/httpbench/)](https://goreportcard.com/report/github.com/rnemeth90/httpbench/)
## Description
HttpBench is a simple utility for bench marking HTTP servers. 

### Dependencies
* to build yourself, you must have Go v1.13+ installed

### Installing
Download the latest release [here](https://github.com/rnemeth90/httpbench/releases)

### Executing program
```
gopher@localhost→ httpbench
httpbench

Usage:
  httpbench --url https://mywebsite.com
  httpbench --url https://mywebsite.com --count 100

Options:
  -b, --bodyFile string   body file in json
  -c, --compression       compression (default true)
  -d, --duration int      duration (default 10)
  -h, --headers string    request headers <string:string>
  -i, --insecure          insecure
  -k, --keepalives        keepalives (default true)
  -r, --requests int      count of requests per second (default 4)
  -t, --timeout int       timeout (default 10)
  -u, --url string        url to test

gopher@localhost→ httpbench www.google.com
Making 4 calls per second for 10 seconds...
Finished sending 4 requests...
Finished sending 8 requests...
Finished sending 12 requests...
Finished sending 16 requests...
Finished sending 20 requests...
Finished sending 24 requests...
Finished sending 28 requests...
Finished sending 32 requests...
Finished sending 36 requests...
Finished sending 40 requests...

----------------------------------
Total Requests: 40
Total Time Taken: 7.098961s
----------------------------------
fastest: 80.2569ms
slowest: 825.6611ms
average: 177.474025ms
----------------------------------
20x count: 40
30x count: 0
40x count: 0
50x count: 0
```

## Help
If you need help or find a bug, submit an issue

## To Do
- [x] Connections
- [x] headers
- [x] url tolower()
- [x] timeout
- [x] statistics
- [x] tests (we don't have 100% code coverage...)
- [x] body
- [ ] finish readme
- [x] support custom ports
- [x] support common HTTP methods
- [x] allow user to specify number of concurrent connections or goroutines
- [x] custom user-agent string
- [ ] change output format (json, csv, etc.)
- [x] proxy support/proxy authentication
- [x] authentication
- [x] cancellation on long requests (using a context)
- [x] skip SSL validation
- [ ] override the host header
- [ ] output version
- [ ] verbose mode(?)
- [ ] improve usage() function

## Version History
* 1.0.0
    * Initial Release

## License
This project is licensed under the MIT License - see the LICENSE.md file for details
