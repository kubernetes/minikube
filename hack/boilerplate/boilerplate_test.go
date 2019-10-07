/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/magiconair/properties/assert"

	"testing"
)

func TestPassContentThesame(t *testing.T) {
	t1 := []string{"this is a just a string"}
	t2 := []string{"this is a just a string"}

	assert.Equal(t, IsContentTheSame(t1, t2), true)
}

func TestFailContentThesame(t *testing.T) {
	t1 := []string{"this is a just a string1"}
	t2 := []string{"this is a just a string2"}

	assert.Equal(t, !IsContentTheSame(t1, t2), true)
}
