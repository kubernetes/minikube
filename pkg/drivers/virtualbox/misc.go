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

package virtualbox

import (
	"bufio"
	"math/rand"
	"os"

	"time"

	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/ssh"
)

// B2DUpdater describes the interactions with b2d.
type B2DUpdater interface {
	UpdateISOCache(storePath, isoURL string) error
	CopyIsoToMachineDir(storePath, machineName, isoURL string) error
}

func NewB2DUpdater() B2DUpdater {
	return &b2dUtilsUpdater{}
}

type b2dUtilsUpdater struct{}

func (u *b2dUtilsUpdater) CopyIsoToMachineDir(storePath, machineName, isoURL string) error {
	return mcnutils.NewB2dUtils(storePath).CopyIsoToMachineDir(isoURL, machineName)
}

func (u *b2dUtilsUpdater) UpdateISOCache(storePath, isoURL string) error {
	return mcnutils.NewB2dUtils(storePath).UpdateISOCache(isoURL)
}

// SSHKeyGenerator describes the generation of ssh keys.
type SSHKeyGenerator interface {
	Generate(path string) error
}

func NewSSHKeyGenerator() SSHKeyGenerator {
	return &defaultSSHKeyGenerator{}
}

type defaultSSHKeyGenerator struct{}

func (g *defaultSSHKeyGenerator) Generate(path string) error {
	return ssh.GenerateSSHKey(path)
}

// LogsReader describes the reading of VBox.log
type LogsReader interface {
	Read(path string) ([]string, error)
}

func NewLogsReader() LogsReader {
	return &vBoxLogsReader{}
}

type vBoxLogsReader struct{}

func (c *vBoxLogsReader) Read(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}

	defer file.Close()

	lines := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

// RandomInter returns random int values.
type RandomInter interface {
	RandomInt(n int) int
}

func NewRandomInter() RandomInter {
	src := rand.NewSource(time.Now().UnixNano())

	return &defaultRandomInter{
		rand: rand.New(src),
	}
}

type defaultRandomInter struct {
	rand *rand.Rand
}

func (d *defaultRandomInter) RandomInt(n int) int {
	return d.rand.Intn(n)
}

// Sleeper sleeps for given duration.
type Sleeper interface {
	Sleep(d time.Duration)
}

func NewSleeper() Sleeper {
	return &defaultSleeper{}
}

type defaultSleeper struct{}

func (s *defaultSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}
