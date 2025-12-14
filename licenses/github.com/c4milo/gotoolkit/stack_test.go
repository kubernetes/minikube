// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gotoolkit

import (
	"testing"
)

func TestPush(t *testing.T) {
	stacks := []Stack{new(ListStack), new(SliceStack)}

	for _, stack := range stacks {
		stack.Push("Test")

		if stack.Size() == 0 {
			t.Errorf("Incorrect result\ngot:  %d\nwant: %d", stack.Size(), 0)
		}
	}

}

func TestPop(t *testing.T) {
	stacks := []Stack{new(ListStack), new(SliceStack)}

	for _, stack := range stacks {
		stack.Push("Test")
		item, err := stack.Pop()

		if err != nil {
			t.Errorf("Incorrect result\ngot:  %v\nwant: %v", err, nil)
		}

		if item != "Test" {
			t.Errorf("Incorrect result\ngot:  %s\nwant: %s", item, "Test")
		}

		if stack.Size() != 0 {
			t.Errorf("Incorrect result\ngot:  %d\nwant: %d", stack.Size(), 0)
		}
	}
}

func TestPopWithEmptyStack(t *testing.T) {
	stacks := []Stack{new(ListStack), new(SliceStack)}

	for _, stack := range stacks {
		item, err := stack.Pop()
		want := "unable to pop element, stack is empty"

		if err.Error() != want {
			t.Errorf("Incorrect result\ngot:  %s\nwant: %s", err, want)
		}

		if item != nil {
			t.Errorf("Incorrect result\ngot:  %v\nwant: %v", item, nil)
		}

		if !stack.IsEmpty() {
			t.Errorf("Incorrect result\ngot:  %t\nwant: %t", stack.IsEmpty(), true)
		}
	}
}

func TestPeek(t *testing.T) {
	stacks := []Stack{new(ListStack), new(SliceStack)}

	for _, stack := range stacks {
		stack.Push("Test")

		item, err := stack.Peek()

		if err != nil {
			t.Errorf("Incorrect result\ngot:  %v\nwant: %v", err, nil)
		}

		if item != "Test" {
			t.Errorf("Incorrect result\ngot:  %s\nwant: %s", item, "Test")
		}

		if stack.Size() != 1 {
			t.Errorf("Incorrect result\ngot:  %d\nwant: %d", stack.Size(), 1)
		}
	}
}
