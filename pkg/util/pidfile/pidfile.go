/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package pidfile

import (
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func GetPids(path string) ([]int, error) {
	out, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "ReadFile")
	}
	klog.Infof("pidfile contents: %s", out)

	pids := []int{}
	strPids := strings.Fields(string(out))
	for _, p := range strPids {
		intPid, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}

		pids = append(pids, intPid)
	}

	return pids, nil
}
