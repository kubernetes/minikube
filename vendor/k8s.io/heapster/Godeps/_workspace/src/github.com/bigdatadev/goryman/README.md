# Riemann client (Golang)

[![GoDoc](https://godoc.org/github.com/bigdatadev/goryman?status.png)](http://godoc.org/github.com/bigdatadev/goryman)

## Introduction

Go client library for [Riemann](https://github.com/aphyr/riemann).

This client library was inspired by [Raidman](https://github.com/amir/raidman), and [Riemann NodeJS Client](https://github.com/perezd/riemann-nodejs-client).

Features:
* Idiomatic concurrency
* Sending events, state updates, queries.
* Feature parity with the reference implementation written in Ruby.

## Installation

To install the package for use in your own programs:

```
go get github.com/bigdatadev/goryman
```

If you're a developer, Riemann uses [Google Protocol Buffers](https://github.com/golang/protobuf), so make sure that's installed and available on your PATH.

```
go get github.com/golang/protobuf/{proto,protoc-gen-go}
```

## Getting Started

First we'll need to import the library:

```go
import (
    "github.com/bigdatadev/goryman"
)
```

Next we'll need to establish a new client:

```go
c := goryman.NewGorymanClient("localhost:5555")
err := c.Connect()
if err != nil {
    panic(err)
}
```

Don't forget to close the client connection when you're done:

```go
defer c.Close()
```

Just like the Riemann Ruby client, the client sends small events over UDP by default. TCP is used for queries, and large events. There is no acknowledgement of UDP packets, but they are roughly an order of magnitude faster than TCP. We assume both TCP and UDP are listening on the same port.

Sending events is easy ([list of valid event properties](http://aphyr.github.com/riemann/concepts.html)):

```go
err = c.SendEvent(&goryman.Event{
    Service: "moargore",
    Metric:  100,
    Tags: []string{"nonblocking"},
})
if err != nil {
    panic(err)
}
```

You can also query events:

```go
events, err := c.QueryEvents("host = \"goryman\"")
if err != nil {
    panic(err)
}
```

The Hostname and Time in events will automatically be replaced with the hostname of the server and the current time if none is specified.

## Integrations

Martini: [GoryMartini](http://github.com/bigdatadev/gorymartini)

## Contributing

Just send me a pull request. Please take a look at the project issues and see how you can help. Here are some tips:
- please add more tests.
- please check your syntax.

## Author

Christopher Gilbert

* Web: [http://cjgilbert.me](http://cjgilbert.me)
* Twitter: [@bigdatadev](https://twitter.com/bigdatadev)
* Linkedin: [/in/christophergilbert](https://www.linkedin.com/in/christophergilbert)

## Copyright

See [LICENSE](LICENSE) document