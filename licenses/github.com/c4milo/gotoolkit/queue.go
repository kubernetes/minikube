// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gotoolkit

import "errors"

// Queue defines an interface for implementing Queue data structures.
type Queue interface {
	Enqueue(item interface{})
	Dequeue() (interface{}, error)
	Peek() (interface{}, error)
	IsEmpty() bool
	Size() uint64
}

// ListQueue implements a queue backed by a linked list, it may be faster than SliceQueue
// but consumes more memory and has worse cache locality than SliceQueue
type ListQueue struct {
	first *node
	last  *node
	size  uint64
}

// Enqueue adds an element to the end of the queue.
func (q *ListQueue) Enqueue(item interface{}) {
	last := new(node)
	last.item = item

	if q.IsEmpty() {
		q.first = last
	} else {
		oldLast := q.last
		oldLast.next = last
	}
	q.last = last
	q.size++
}

// Dequeue removes the first element from the queue.
func (q *ListQueue) Dequeue() (interface{}, error) {
	if q.IsEmpty() {
		q.last = nil
		return nil, errors.New("unable to dequeue element, queue is empty")
	}

	item := q.first.item
	q.first = q.first.next
	q.size--
	return item, nil
}

// Peek returns the first element in the queue without removing it.
func (q *ListQueue) Peek() (interface{}, error) {
	if q.IsEmpty() {
		return nil, errors.New("unable to peek element, queue is empty")
	}
	return q.first.item, nil
}

// IsEmpty returns whether the queue is empty or not.
func (q *ListQueue) IsEmpty() bool {
	return q.first == nil
}

// Size returns the number of elements in the queue.
func (q *ListQueue) Size() uint64 {
	return q.size
}

// SliceQueue implements a queue backed by a growing slice. Useful for memory
// constrained environments. It also has better cache locality than ListQueue
type SliceQueue struct {
	size    uint64
	backing []interface{}
}

// Enqueue adds an element to the end of the queue.
func (s *SliceQueue) Enqueue(item interface{}) {
	s.size++
	s.backing = append(s.backing, item)
}

// Peek returns the first element of the queue without removing it from the queue.
func (s *SliceQueue) Peek() (interface{}, error) {
	if s.IsEmpty() {
		return nil, errors.New("unable to peek element, queue is empty")
	}
	return s.backing[0], nil
}

// Dequeue removes and return the first element from the queue.
func (s *SliceQueue) Dequeue() (interface{}, error) {
	if s.IsEmpty() {
		return nil, errors.New("unable to dequeue element, queue is empty")
	}

	item := s.backing[0]
	// https://github.com/golang/go/wiki/SliceTricks
	s.backing = append(s.backing[:0], s.backing[1:]...)
	s.size--
	return item, nil
}

// IsEmpty returns whether or not the queue is empty.
func (s *SliceQueue) IsEmpty() bool {
	return len(s.backing) == 0
}

// Size returns the number of elements in the queue.
func (s *SliceQueue) Size() uint64 {
	return s.size
}
