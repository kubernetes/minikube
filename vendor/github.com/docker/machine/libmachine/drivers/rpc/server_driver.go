package rpcdriver

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/version"
)

type Stacker interface {
	Stack() []byte
}

type StandardStack struct{}

func (ss *StandardStack) Stack() []byte {
	return debug.Stack()
}

var (
	stdStacker Stacker = &StandardStack{}
)

func init() {
	gob.Register(new(RPCFlags))
	gob.Register(new(mcnflag.IntFlag))
	gob.Register(new(mcnflag.StringFlag))
	gob.Register(new(mcnflag.StringSliceFlag))
	gob.Register(new(mcnflag.BoolFlag))
}

type RPCFlags struct {
	Values map[string]interface{}
}

func (r RPCFlags) Get(key string) interface{} {
	val, ok := r.Values[key]
	if !ok {
		log.Warnf("Trying to access option %s which does not exist", key)
		log.Warn("THIS ***WILL*** CAUSE UNEXPECTED BEHAVIOR")
	}
	return val
}

func (r RPCFlags) String(key string) string {
	val, ok := r.Get(key).(string)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to string for key %s", key)
	}
	return val
}

func (r RPCFlags) StringSlice(key string) []string {
	val, ok := r.Get(key).([]string)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to string slice for key %s", key)
	}
	return val
}

func (r RPCFlags) Int(key string) int {
	val, ok := r.Get(key).(int)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to int for key %s", key)
	}
	return val
}

func (r RPCFlags) Bool(key string) bool {
	val, ok := r.Get(key).(bool)
	if !ok {
		log.Warnf("Type assertion did not go smoothly to bool for key %s", key)
	}
	return val
}

type RPCServerDriver struct {
	ActualDriver drivers.Driver
	CloseCh      chan bool
	HeartbeatCh  chan bool
}

func NewRPCServerDriver(d drivers.Driver) *RPCServerDriver {
	return &RPCServerDriver{
		ActualDriver: d,
		CloseCh:      make(chan bool),
		HeartbeatCh:  make(chan bool),
	}
}

func (r *RPCServerDriver) Close(_, _ *struct{}) error {
	r.CloseCh <- true
	return nil
}

func (r *RPCServerDriver) GetVersion(_ *struct{}, reply *int) error {
	*reply = version.APIVersion
	return nil
}

func (r *RPCServerDriver) GetConfigRaw(_ *struct{}, reply *[]byte) error {
	driverData, err := json.Marshal(r.ActualDriver)
	if err != nil {
		return err
	}

	*reply = driverData

	return nil
}

func (r *RPCServerDriver) GetCreateFlags(_ *struct{}, reply *[]mcnflag.Flag) error {
	*reply = r.ActualDriver.GetCreateFlags()
	return nil
}

func (r *RPCServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	return json.Unmarshal(data, &r.ActualDriver)
}

func trapPanic(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("Panic in the driver: %s\n%s", r.(error), stdStacker.Stack())
	}
}

func (r *RPCServerDriver) Create(_, _ *struct{}) (err error) {
	// In an ideal world, plugins wouldn't ever panic.  However, panics
	// have been known to happen and cause issues.  Therefore, we recover
	// and do not crash the RPC server completely in the case of a panic
	// during create.
	defer trapPanic(&err)

	err = r.ActualDriver.Create()

	return err
}

func (r *RPCServerDriver) DriverName(_ *struct{}, reply *string) error {
	*reply = r.ActualDriver.DriverName()
	return nil
}

func (r *RPCServerDriver) GetIP(_ *struct{}, reply *string) error {
	ip, err := r.ActualDriver.GetIP()
	*reply = ip
	return err
}

func (r *RPCServerDriver) GetMachineName(_ *struct{}, reply *string) error {
	*reply = r.ActualDriver.GetMachineName()
	return nil
}

func (r *RPCServerDriver) GetSSHHostname(_ *struct{}, reply *string) error {
	hostname, err := r.ActualDriver.GetSSHHostname()
	*reply = hostname
	return err
}

func (r *RPCServerDriver) GetSSHKeyPath(_ *struct{}, reply *string) error {
	*reply = r.ActualDriver.GetSSHKeyPath()
	return nil
}

// GetSSHPort returns port for use with ssh
func (r *RPCServerDriver) GetSSHPort(_ *struct{}, reply *int) error {
	port, err := r.ActualDriver.GetSSHPort()
	*reply = port
	return err
}

func (r *RPCServerDriver) GetSSHUsername(_ *struct{}, reply *string) error {
	*reply = r.ActualDriver.GetSSHUsername()
	return nil
}

func (r *RPCServerDriver) GetURL(_ *struct{}, reply *string) error {
	info, err := r.ActualDriver.GetURL()
	*reply = info
	return err
}

func (r *RPCServerDriver) GetState(_ *struct{}, reply *state.State) error {
	s, err := r.ActualDriver.GetState()
	*reply = s
	return err
}

func (r *RPCServerDriver) Kill(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Kill()
}

func (r *RPCServerDriver) PreCreateCheck(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.PreCreateCheck()
}

func (r *RPCServerDriver) Remove(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Remove()
}

func (r *RPCServerDriver) Restart(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Restart()
}

func (r *RPCServerDriver) SetConfigFromFlags(flags *drivers.DriverOptions, _ *struct{}) error {
	return r.ActualDriver.SetConfigFromFlags(*flags)
}

func (r *RPCServerDriver) Start(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Start()
}

func (r *RPCServerDriver) Stop(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Stop()
}

func (r *RPCServerDriver) Heartbeat(_ *struct{}, _ *struct{}) error {
	r.HeartbeatCh <- true
	return nil
}
