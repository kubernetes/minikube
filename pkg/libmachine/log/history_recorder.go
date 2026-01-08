/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package log

import (
	"fmt"
	"sync"
)

type HistoryRecorder struct {
	lock    *sync.Mutex
	records []string
}

func NewHistoryRecorder() *HistoryRecorder {
	return &HistoryRecorder{
		lock:    &sync.Mutex{},
		records: []string{},
	}
}

func (ml *HistoryRecorder) History() []string {
	return ml.records
}

func (ml *HistoryRecorder) Record(args ...interface{}) {
	ml.lock.Lock()
	defer ml.lock.Unlock()
	ml.records = append(ml.records, fmt.Sprint(args...))
}

func (ml *HistoryRecorder) Recordf(fmtString string, args ...interface{}) {
	ml.lock.Lock()
	defer ml.lock.Unlock()
	ml.records = append(ml.records, fmt.Sprintf(fmtString, args...))
}
