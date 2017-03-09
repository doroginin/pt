# pt

This software `pt` (performance testing) is a HTTP load testing tool.

## Install
    go get github.com/rilinor/pt...


## Usage

    pt -url http://hostname[:port][/path] [options]

    Options are:
        -help
            Show help
        -c int
            Number of multiple requests to make at a time (default 50)
        -cert string
            TLS client PEM encoded certificate file
        -ka
            Use keep alive
        -n int
            The total number of requests (0 - unlimit)
        -rt duration
            Run time of benchmark (0 - unlimit)
        -t duration
            Timeout per request (default 5s)
        -url string
            Fetching url

## Examples

    pt \
        -url http://localhost:1234 \ # fetching url
        -ka \                        # use keep-alive
        -n 1000000 \                 # total number of requests
        -c 10000 \                   # concurrency level
        -t 1m                        # timeout per request
