package kafka

import (
	"errors"
	"sync"

	"github.com/optiopay/kafka/proto"
)

// ErrMxClosed is returned as a result of closed multiplexer consumption.
var ErrMxClosed = errors.New("closed")

// Mx is multiplexer combining into single stream number of consumers.
//
// It is responsibility of the user of the multiplexer and the consumer
// implementation to handle errors.
// ErrNoData returned by consumer is not passed through by the multiplexer,
// instead consumer that returned ErrNoData is removed from merged set. When
// all consumers are removed (set is empty), Mx is automatically closed and any
// further Consume call will result in ErrMxClosed error.
//
// It is important to remember that because fetch from every consumer is done
// by separate worker, most of the time there is one message consumed by each
// worker that is held in memory while waiting for opportunity to return it
// once Consume on multiplexer is called. Closing multiplexer may result in
// ignoring some of already read, waiting for delivery messages kept internally
// by every worker.
type Mx struct {
	errc chan error
	msgc chan *proto.Message
	stop chan struct{}

	mu      sync.Mutex
	closed  bool
	workers int
}

// Merge is merging consume result of any number of consumers into single stream
// and expose them through returned multiplexer.
func Merge(consumers ...Consumer) *Mx {
	p := &Mx{
		errc:    make(chan error),
		msgc:    make(chan *proto.Message),
		stop:    make(chan struct{}),
		workers: len(consumers),
	}

	for _, consumer := range consumers {
		go func(c Consumer) {
			defer func() {
				p.mu.Lock()
				p.workers -= 1
				if p.workers == 0 && !p.closed {
					close(p.stop)
					p.closed = true
				}
				p.mu.Unlock()
			}()

			for {
				msg, err := c.Consume()
				if err != nil {
					if err == ErrNoData {
						return
					}
					select {
					case p.errc <- err:
					case <-p.stop:
						return
					}
				} else {
					select {
					case p.msgc <- msg:
					case <-p.stop:
						return
					}
				}
			}
		}(consumer)
	}

	return p
}

// Workers return number of active consumer workers that are pushing messages
// to multiplexer conumer queue.
func (p *Mx) Workers() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.workers
}

// Close is closing multiplexer and stopping all underlying workers.
//
// Closing multiplexer will stop all workers as soon as possible, but any
// consume-in-progress action performed by worker has to be finished first. Any
// consumption result received after closing multiplexer is ignored.
//
// Close is returning without waiting for all the workers to finish.
//
// Closing closed multiplexer has no effect.
func (p *Mx) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.closed {
		p.closed = true
		close(p.stop)
	}
}

// Consume returns Consume result from any of the merged consumer.
func (p *Mx) Consume() (*proto.Message, error) {
	select {
	case <-p.stop:
		return nil, ErrMxClosed
	case msg := <-p.msgc:
		return msg, nil
	case err := <-p.errc:
		return nil, err
	}
}
