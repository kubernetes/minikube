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

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

type resultManager struct {
	results map[*Binary]resultWrapper
}

func newResultManager() *resultManager {
	return &resultManager{
		results: map[*Binary]resultWrapper{},
	}
}

func (rm *resultManager) addResult(binary *Binary, test string, r result) {
	a, ok := rm.results[binary]
	if !ok {
		r := map[string][]*result{test: {&r}}
		rm.results[binary] = resultWrapper{r}
		return
	}
	b, ok := a.results[test]
	if !ok {
		a.results[test] = []*result{&r}
		return
	}
	a.results[test] = append(b, &r)
}

func (rm *resultManager) totalTimes(binary *Binary, t string) []float64 {
	var totals []float64
	results, ok := rm.results[binary].results[t]
	if !ok {
		return nil
	}
	for _, r := range results {
		total := 0.0
		for _, t := range r.timedLogs {
			total += t
		}
		totals = append(totals, total)
	}
	return totals
}

func (rm *resultManager) summarizeResults(binaries []*Binary) {
	// print total and average times
	table := make([][]string, 2)
	for i := range table {
		table[i] = make([]string, len(binaries)+1)
	}
	table[0][0] = "minikube start"
	table[1][0] = "enable ingress"
	totalTimes := make(map[string]map[string][]float64)
	for i := range rm.results[binaries[0]].results {
		totalTimes[i] = make(map[string][]float64)
	}

	for i, b := range binaries {
		for t := range rm.results[b].results {
			index := 0
			if t == "ingress" {
				index = 1
			}
			totalTimes[t][b.Name()] = rm.totalTimes(b, t)
			table[index][i+1] = fmt.Sprintf("%.1fs", average(totalTimes[t][b.Name()]))
		}
	}
	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Command", binaries[0].Name(), binaries[1].Name()})
	for _, v := range table {
		// Add warning sign if PR average is 5 seconds higher than average at HEAD
		if len(v) > 3 {
			prTime, _ := strconv.ParseFloat(v[2][:len(v[2])-1], 64)
			headTime, _ := strconv.ParseFloat(v[1][:len(v[1])-1], 64)
			if prTime-headTime > threshold {
				v[0] = fmt.Sprintf("⚠️  %s", v[0])
				v[2] = fmt.Sprintf("%s ⚠️", v[2])
			}
		}
		t.Append(v)
	}
	fmt.Println("```")
	t.Render()
	fmt.Println("```")
	fmt.Println()

	fmt.Println("<details>")
	fmt.Println()
	for t, times := range totalTimes {
		for b, f := range times {
			fmt.Printf("Times for %s %s: ", b, t)
			for _, tt := range f {
				fmt.Printf("%.1fs ", tt)
			}
			fmt.Println()
		}
		fmt.Println()
	}
	fmt.Println("</details>")
}
