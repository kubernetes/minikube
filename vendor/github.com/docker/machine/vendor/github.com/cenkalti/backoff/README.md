# backoff

[![GoDoc](https://godoc.org/github.com/cenkalti/backoff?status.png)](https://godoc.org/github.com/cenkalti/backoff)
[![Build Status](https://travis-ci.org/cenkalti/backoff.png)](https://travis-ci.org/cenkalti/backoff)

This is a Go port of the exponential backoff algorithm from
[google-http-java-client](https://code.google.com/p/google-http-java-client/wiki/ExponentialBackoff).

[Exponential backoff](http://en.wikipedia.org/wiki/Exponential_backoff)
is an algorithm that uses feedback to multiplicatively decrease the rate of some process,
in order to gradually find an acceptable rate.
The retries exponentially increase and stop increasing when a certain threshold is met.




## Install

```bash
go get github.com/cenkalti/backoff
```

## Example

Simple retry helper that uses exponential back-off algorithm:

```go
operation := func() error {
    // An operation that might fail
}

err := backoff.Retry(operation, backoff.NewExponentialBackOff())
if err != nil {
    // handle error
}

// operation is successfull
```

Ticker example:

```go
operation := func() error {
    // An operation that may fail
}

b := backoff.NewExponentialBackOff()
ticker := backoff.NewTicker(b)

var err error

// Ticks will continue to arrive when the previous operation is still running,
// so operations that take a while to fail could run in quick succession.
for t = range ticker.C {
    if err = operation(); err != nil {
        log.Println(err, "will retry...")
        continue
    }

    ticker.Stop()
    break
}

if err != nil {
    // Operation has failed.
}

// Operation is successfull.
```
