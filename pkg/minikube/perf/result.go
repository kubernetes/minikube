/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package perf

type resultWrapper struct {
	results map[string][]*result
}

type result struct {
	logs      []string
	timedLogs map[string]float64
}

func newResult() *result {
	return &result{
		timedLogs: map[string]float64{},
	}
}

func (r *result) addTimedLog(log string, time float64) {
	r.logs = append(r.logs, log)
	r.timedLogs[log] = time
}
