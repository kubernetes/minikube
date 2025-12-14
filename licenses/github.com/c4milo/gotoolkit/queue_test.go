// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gotoolkit

import (
	"testing"
)

func TestEnqueue(t *testing.T) {
	queues := []Queue{new(ListQueue), new(SliceQueue)}

	for _, queue := range queues {
		queue.Enqueue("Test")

		if queue.Size() == 0 {
			t.Errorf("Incorrect result\ngot:  %d\nwant: %d", queue.Size(), 0)
		}
	}

}

func TestDequeue(t *testing.T) {
	queues := []Queue{new(ListQueue), new(SliceQueue)}

	for _, queue := range queues {
		queue.Enqueue("Test1")
		queue.Enqueue("Test2")
		queue.Enqueue("Test3")
		item, err := queue.Dequeue()

		if err != nil {
			t.Errorf("Incorrect result\ngot:  %v\nwant: %v", err, nil)
		}

		if item != "Test1" {
			t.Errorf("Incorrect result\ngot:  %s\nwant: %s", item, "Test1")
		}

		if queue.Size() != 2 {
			t.Errorf("Incorrect result\ngot:  %d\nwant: %d", queue.Size(), 2)
		}
	}
}

func TestDequeueWithEmptyQueue(t *testing.T) {
	queues := []Queue{new(ListQueue), new(SliceQueue)}

	for _, queue := range queues {
		item, err := queue.Dequeue()
		want := "unable to dequeue element, queue is empty"

		if err.Error() != want {
			t.Errorf("Incorrect result\ngot:  %s\nwant: %s", err, want)
		}

		if item != nil {
			t.Errorf("Incorrect result\ngot:  %v\nwant: %v", item, nil)
		}

		if !queue.IsEmpty() {
			t.Errorf("Incorrect result\ngot:  %t\nwant: %t", queue.IsEmpty(), true)
		}
	}
}

func TestPeekQueue(t *testing.T) {
	queues := []Queue{new(ListQueue), new(SliceQueue)}

	for _, queue := range queues {
		queue.Enqueue("Test")

		item, err := queue.Peek()

		if err != nil {
			t.Errorf("Incorrect result\ngot:  %v\nwant: %v", err, nil)
		}

		if item != "Test" {
			t.Errorf("Incorrect result\ngot:  %s\nwant: %s", item, "Test")
		}

		if queue.Size() != 1 {
			t.Errorf("Incorrect result\ngot:  %d\nwant: %d", queue.Size(), 1)
		}
	}
}
