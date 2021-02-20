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
	"strings"
	"testing"
)

func TestNumaXml(t *testing.T) {
	_, err := numaXML(1, 1024, 0)
	if err == nil {
		t.Errorf("check invalid numa count failed: %s", err)
	}

	xml, err := numaXML(10, 10240, 8)
	expXML := `<numa>
  <cell id='0' cpus='0,1' memory='1280' unit='MiB'/>
  <cell id='1' cpus='2,3' memory='1280' unit='MiB'/>
  <cell id='2' cpus='4' memory='1280' unit='MiB'/>
  <cell id='3' cpus='5' memory='1280' unit='MiB'/>
  <cell id='4' cpus='6' memory='1280' unit='MiB'/>
  <cell id='5' cpus='7' memory='1280' unit='MiB'/>
  <cell id='6' cpus='8' memory='1280' unit='MiB'/>
  <cell id='7' cpus='9' memory='1280' unit='MiB'/>
</numa>`
	if err != nil {
		t.Errorf("gen xml failed: %s", err)
	}
	if strings.TrimSpace(xml) != expXML {
		t.Errorf("gen xml: %s not match expect xml: %s", xml, expXML)
	}

}
