/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package kvm

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

// NUMATmpl NUMA XML Template
const NUMATmpl = `
<numa>
  {{- range $idx,$val :=. }}
  <cell id='{{$idx}}' cpus='{{$val.CPUTopology}}' memory='{{$val.Memory}}' unit='MiB'/>
  {{- end }}
</numa>
`

// NUMA this struct use for NUMATmpl
type NUMA struct {
	CPUCount    int
	Memory      int
	CPUTopology string
}

// NumaXml generate numa xml
// evenly distributed cpu core & memory to each numa node
func NumaXml(cpu, memory, numaCount int) (string, error) {
	if numaCount < 1 {
		return "", fmt.Errorf("numa node count must >= 1")
	}
	if cpu < numaCount {
		return "", fmt.Errorf("cpu count must >= numa node count")
	}
	numaNodes := make([]*NUMA, numaCount)
	CPUSeq := 0
	cpuBaseCount := cpu / numaCount
	cpuExtraCount := cpu % numaCount

	for i := range numaNodes {
		numaNodes[i] = &NUMA{CPUCount: cpuBaseCount}
	}

	for i := 0; i < cpuExtraCount; i++ {
		numaNodes[i].CPUCount++
	}
	for i := range numaNodes {
		CPUTopologySlice := make([]string, 0)
		for seq := CPUSeq; seq < CPUSeq+numaNodes[i].CPUCount; seq++ {
			CPUTopologySlice = append(CPUTopologySlice, strconv.Itoa(seq))
		}
		numaNodes[i].CPUTopology = strings.Join(CPUTopologySlice, ",")
		CPUSeq += numaNodes[i].CPUCount
	}

	memoryBaseCount := memory / numaCount
	memoryExtraCount := memory % numaCount

	for i := range numaNodes {
		numaNodes[i].Memory = memoryBaseCount
	}

	for i := 0; i < memoryExtraCount; i++ {
		numaNodes[i].Memory++
	}

	tmpl := template.Must(template.New("numa").Parse(NUMATmpl))
	var NUMAXML bytes.Buffer
	if err := tmpl.Execute(&NUMAXML, numaNodes); err != nil {
		return "", fmt.Errorf("couldn't generate numa XML: %v", err)
	}
	return NUMAXML.String(), nil
}
