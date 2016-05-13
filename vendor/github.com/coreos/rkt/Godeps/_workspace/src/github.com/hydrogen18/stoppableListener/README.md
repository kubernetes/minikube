stoppableListener
=================

An example of a stoppable TCP listener in Go. This library wraps an existing TCP connection object. A goroutine calling `Accept()`
is interrupted with `StoppedError` whenever the listener is stopped by a call to `Stop()`. Usage is demonstrated below, and in `example/example.go`.


```
	originalListener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		panic(err)
	}
```
