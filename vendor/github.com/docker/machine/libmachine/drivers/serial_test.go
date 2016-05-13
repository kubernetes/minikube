package drivers

import (
	"testing"

	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
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

func (d *MockDriver) Create() error {
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

func (d *MockDriver) GetState() (state.State, error) {
	d.calls.record("GetState")
	return d.state, nil
}

func (d *MockDriver) Kill() error {
	d.calls.record("Kill")
	return nil
}

func (d *MockDriver) PreCreateCheck() error {
	d.calls.record("PreCreateCheck")
	return nil
}

func (d *MockDriver) Remove() error {
	d.calls.record("Remove")
	return nil
}

func (d *MockDriver) Restart() error {
	d.calls.record("Restart")
	return nil
}

func (d *MockDriver) SetConfigFromFlags(opts DriverOptions) error {
	d.calls.record("SetConfigFromFlags")
	return nil
}

func (d *MockDriver) Start() error {
	d.calls.record("Start")
	return nil
}

func (d *MockDriver) Stop() error {
	d.calls.record("Stop")
	return nil
}

func TestSerialDriverCreate(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	err := driver.Create()

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
	machineState, _ := driver.GetState()

	assert.Equal(t, state.Running, machineState)
	assert.Equal(t, []string{"Lock", "GetState", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverKill(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	driver.Kill()

	assert.Equal(t, []string{"Lock", "Kill", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverPreCreateCheck(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	driver.PreCreateCheck()

	assert.Equal(t, []string{"Lock", "PreCreateCheck", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverRemove(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	driver.Remove()

	assert.Equal(t, []string{"Lock", "Remove", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverRestart(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	driver.Restart()

	assert.Equal(t, []string{"Lock", "Restart", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverSetConfigFromFlags(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	driver.SetConfigFromFlags(nil)

	assert.Equal(t, []string{"Lock", "SetConfigFromFlags", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverStart(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	driver.Start()

	assert.Equal(t, []string{"Lock", "Start", "Unlock"}, callRecorder.calls)
}

func TestSerialDriverStop(t *testing.T) {
	callRecorder := &CallRecorder{}

	driver := newSerialDriverWithLock(&MockDriver{calls: callRecorder}, &MockLocker{calls: callRecorder})
	driver.Stop()

	assert.Equal(t, []string{"Lock", "Stop", "Unlock"}, callRecorder.calls)
}
