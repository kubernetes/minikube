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

package drivers

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
)

type CallRecorder struct {
	calls []string
}

func (c *CallRecorder) record(call string) {
	c.calls = append(c.calls, call)
}

type MockLocker struct {
	calls *CallRecorder
}

func (l *MockLocker) Lock() {
	l.calls.record("Lock")
}

func (l *MockLocker) Unlock() {
	l.calls.record("Unlock")
}

// x7TODO: another mock driver.. There's like 3 of them so far..
// should be able to merge those into 1 driver inside libmachine/drivers
type MockDriver struct {
	calls       *CallRecorder
	driverName  string
	flags       []mcnflag.Flag
	ip          string
	machineName string
	sshHostname string
	sshKeyPath  string
	sshPort     int
	sshUsername string
	url         string
	state       state.State
}

func (d *MockDriver) GetRunner() (runner.Runner, error) {
	d.calls.record("GetRunner")
	return nil, nil
}

func (d *MockDriver) RunCmd(_ *exec.Cmd) (*runner.RunResult, error) {
	d.calls.record("RunCmd")
	return nil, nil
}

func (d *MockDriver) IsContainerBased() bool {
	d.calls.record("IsContainerBased")
	return false
}

func (d *MockDriver) IsISOBased() bool {
	d.calls.record("IsISOBased")
	return false
}

func (d *MockDriver) IsManaged() bool {
	d.calls.record("IsManaged")
	return false
}

func (d *MockDriver) CreateMachine() error {
	d.calls.record("Create")
	return nil
}

func (d *MockDriver) DriverName() string {
	d.calls.record("DriverName")
	return d.driverName
}

func (d *MockDriver) GetCreateFlags() []mcnflag.Flag {
	d.calls.record("GetCreateFlags")
	return d.flags
}

func (d *MockDriver) GetIP() (string, error) {
	d.calls.record("GetIP")
	return d.ip, nil
}

func (d *MockDriver) GetMachineName() string {
	d.calls.record("GetMachineName")
	return d.machineName
}

func (d *MockDriver) GetSSHHostname() (string, error) {
	d.calls.record("GetSSHHostname")
	return d.sshHostname, nil
}

func (d *MockDriver) GetSSHKeyPath() string {
	d.calls.record("GetSSHKeyPath")
	return d.sshKeyPath
}

func (d *MockDriver) GetSSHPort() (int, error) {
	d.calls.record("GetSSHPort")
	return d.sshPort, nil
}

func (d *MockDriver) GetSSHUsername() string {
	d.calls.record("GetSSHUsername")
	return d.sshUsername
}

func (d *MockDriver) GetURL() (string, error) {
	d.calls.record("GetURL")
	return d.url, nil
}

func (d *MockDriver) GetMachineState() (state.State, error) {
	d.calls.record("GetMachineState")
	return d.state, nil
}

func (d *MockDriver) KillMachine() error {
	d.calls.record("KillMachine")
	return nil
}

func (d *MockDriver) PreCreateCheck() error {
	d.calls.record("PreCreateCheck")
	return nil
}

func (d *MockDriver) RemoveMachine() error {
	d.calls.record("RemoveMachine")
	return nil
}

func (d *MockDriver) RestartMachine() error {
	d.calls.record("RestartMachine")
	return nil
}

func (d *MockDriver) SetConfigFromFlags(_ DriverOptions) error {
	d.calls.record("SetConfigFromFlags")
	return nil
}

func (d *MockDriver) StartMachine() error {
	d.calls.record("StartMachine")
	return nil
}

func (d *MockDriver) StopMachine() error {
	d.calls.record("StopMachine")
	return nil
}

func TestSerialDriverCreate(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	err := driver.CreateMachine()

	assert.NoError(t, err)
	assert.Equal(t, []string{"Lock", "Create", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverDriverName(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{driverName: "DRIVER", calls: callRecorder}, &MockLocker{calls: callRecorder})
	driverName := driver.DriverName()

	assert.Equal(t, "DRIVER", driverName)
	assert.Equal(t, []string{"Lock", "DriverName", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetIP(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{ip: "IP", calls: callRecorder}, &MockLocker{calls: callRecorder})
	ip, _ := driver.GetIP()

	assert.Equal(t, "IP", ip)
	assert.Equal(t, []string{"Lock", "GetIP", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetMachineName(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{machineName: "MACHINE_NAME", calls: callRecorder}, &MockLocker{calls: callRecorder})
	machineName := driver.GetMachineName()

	assert.Equal(t, "MACHINE_NAME", machineName)
	assert.Equal(t, []string{"Lock", "GetMachineName", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetSSHHostname(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{sshHostname: "SSH_HOSTNAME", calls: callRecorder}, &MockLocker{calls: callRecorder})
	sshHostname, _ := driver.GetSSHHostname()

	assert.Equal(t, "SSH_HOSTNAME", sshHostname)
	assert.Equal(t, []string{"Lock", "GetSSHHostname", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetSSHKeyPath(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{sshKeyPath: "PATH", calls: callRecorder}, &MockLocker{calls: callRecorder})
	path := driver.GetSSHKeyPath()

	assert.Equal(t, "PATH", path)
	assert.Equal(t, []string{"Lock", "GetSSHKeyPath", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetSSHPort(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{sshPort: 42, calls: callRecorder}, &MockLocker{calls: callRecorder})
	sshPort, _ := driver.GetSSHPort()

	assert.Equal(t, 42, sshPort)
	assert.Equal(t, []string{"Lock", "GetSSHPort", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetSSHUsername(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{sshUsername: "SSH_USER", calls: callRecorder}, &MockLocker{calls: callRecorder})
	sshUsername := driver.GetSSHUsername()

	assert.Equal(t, "SSH_USER", sshUsername)
	assert.Equal(t, []string{"Lock", "GetSSHUsername", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetURL(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{url: "URL", calls: callRecorder}, &MockLocker{calls: callRecorder})
	url, _ := driver.GetURL()

	assert.Equal(t, "URL", url)
	assert.Equal(t, []string{"Lock", "GetURL", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverGetState(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{state: state.Running, calls: callRecorder}, &MockLocker{calls: callRecorder})
	machineState, _ := driver.GetMachineState()

	assert.Equal(t, state.Running, machineState)
	assert.Equal(t, []string{"Lock", "GetState", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverKill(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	_ = driver.KillMachine()

	assert.Equal(t, []string{"Lock", "Kill", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverPreCreateCheck(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	_ = driver.PreCreateCheck()

	assert.Equal(t, []string{"Lock", "PreCreateCheck", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverRemove(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	_ = driver.RemoveMachine()

	assert.Equal(t, []string{"Lock", "Remove", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverRestart(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	_ = driver.RestartMachine()

	assert.Equal(t, []string{"Lock", "Restart", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverSetConfigFromFlags(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	_ = driver.SetConfigFromFlags(nil)

	assert.Equal(t, []string{"Lock", "SetConfigFromFlags", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverStart(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	_ = driver.StartMachine()

	assert.Equal(t, []string{"Lock", "Start", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverStop(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	_ = driver.StopMachine()

	assert.Equal(t, []string{"Lock", "Stop", "Unlock"}, callRecorder.calls)
}
