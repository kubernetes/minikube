package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"errors"

	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/crashreport"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/swarm"
)

var (
	errNoMachineName = errors.New("Error: No machine name specified")
)

var (
	SharedCreateFlags = []cli.Flag{
		cli.StringFlag{
			Name: "driver, d",
			Usage: fmt.Sprintf(
				"Driver to create machine with.",
			),
			Value:  "none",
			EnvVar: "MACHINE_DRIVER",
		},
		cli.StringFlag{
			Name:   "engine-install-url",
			Usage:  "Custom URL to use for engine installation",
			Value:  drivers.DefaultEngineInstallURL,
			EnvVar: "MACHINE_DOCKER_INSTALL_URL",
		},
		cli.StringSliceFlag{
			Name:  "engine-opt",
			Usage: "Specify arbitrary flags to include with the created engine in the form flag=value",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "engine-insecure-registry",
			Usage: "Specify insecure registries to allow with the created engine",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:   "engine-registry-mirror",
			Usage:  "Specify registry mirrors to use",
			Value:  &cli.StringSlice{},
			EnvVar: "ENGINE_REGISTRY_MIRROR",
		},
		cli.StringSliceFlag{
			Name:  "engine-label",
			Usage: "Specify labels for the created engine",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "engine-storage-driver",
			Usage: "Specify a storage driver to use with the engine",
		},
		cli.StringSliceFlag{
			Name:  "engine-env",
			Usage: "Specify environment variables to set in the engine",
			Value: &cli.StringSlice{},
		},
		cli.BoolFlag{
			Name:  "swarm",
			Usage: "Configure Machine to join a Swarm cluster",
		},
		cli.StringFlag{
			Name:   "swarm-image",
			Usage:  "Specify Docker image to use for Swarm",
			Value:  "swarm:latest",
			EnvVar: "MACHINE_SWARM_IMAGE",
		},
		cli.BoolFlag{
			Name:  "swarm-master",
			Usage: "Configure Machine to be a Swarm master",
		},
		cli.StringFlag{
			Name:  "swarm-discovery",
			Usage: "Discovery service to use with Swarm",
			Value: "",
		},
		cli.StringFlag{
			Name:  "swarm-strategy",
			Usage: "Define a default scheduling strategy for Swarm",
			Value: "spread",
		},
		cli.StringSliceFlag{
			Name:  "swarm-opt",
			Usage: "Define arbitrary flags for Swarm master",
			Value: &cli.StringSlice{},
		},
		cli.StringSliceFlag{
			Name:  "swarm-join-opt",
			Usage: "Define arbitrary flags for Swarm join",
			Value: &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:  "swarm-host",
			Usage: "ip/socket to listen on for Swarm master",
			Value: "tcp://0.0.0.0:3376",
		},
		cli.StringFlag{
			Name:  "swarm-addr",
			Usage: "addr to advertise for Swarm (default: detect and use the machine IP)",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "swarm-experimental",
			Usage: "Enable Swarm experimental features",
		},
		cli.StringSliceFlag{
			Name:  "tls-san",
			Usage: "Support extra SANs for TLS certs",
			Value: &cli.StringSlice{},
		},
	}
)

func cmdCreateInner(c CommandLine, api libmachine.API) error {
	if len(c.Args()) > 1 {
		return fmt.Errorf("Invalid command line. Found extra arguments %v", c.Args()[1:])
	}

	name := c.Args().First()
	if name == "" {
		c.ShowHelp()
		return errNoMachineName
	}

	validName := host.ValidateHostName(name)
	if !validName {
		return fmt.Errorf("Error creating machine: %s", mcnerror.ErrInvalidHostname)
	}

	if err := validateSwarmDiscovery(c.String("swarm-discovery")); err != nil {
		return fmt.Errorf("Error parsing swarm discovery: %s", err)
	}

	// TODO: Fix hacky JSON solution
	rawDriver, err := json.Marshal(&drivers.BaseDriver{
		MachineName: name,
		StorePath:   c.GlobalString("storage-path"),
	})
	if err != nil {
		return fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	driverName := c.String("driver")
	h, err := api.NewHost(driverName, rawDriver)
	if err != nil {
		return fmt.Errorf("Error getting new host: %s", err)
	}

	h.HostOptions = &host.Options{
		AuthOptions: &auth.Options{
			CertDir:          mcndirs.GetMachineCertDir(),
			CaCertPath:       tlsPath(c, "tls-ca-cert", "ca.pem"),
			CaPrivateKeyPath: tlsPath(c, "tls-ca-key", "ca-key.pem"),
			ClientCertPath:   tlsPath(c, "tls-client-cert", "cert.pem"),
			ClientKeyPath:    tlsPath(c, "tls-client-key", "key.pem"),
			ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), name, "server.pem"),
			ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), name, "server-key.pem"),
			StorePath:        filepath.Join(mcndirs.GetMachineDir(), name),
			ServerCertSANs:   c.StringSlice("tls-san"),
		},
		EngineOptions: &engine.Options{
			ArbitraryFlags:   c.StringSlice("engine-opt"),
			Env:              c.StringSlice("engine-env"),
			InsecureRegistry: c.StringSlice("engine-insecure-registry"),
			Labels:           c.StringSlice("engine-label"),
			RegistryMirror:   c.StringSlice("engine-registry-mirror"),
			StorageDriver:    c.String("engine-storage-driver"),
			TLSVerify:        true,
			InstallURL:       c.String("engine-install-url"),
		},
		SwarmOptions: &swarm.Options{
			IsSwarm:            c.Bool("swarm") || c.Bool("swarm-master"),
			Image:              c.String("swarm-image"),
			Agent:              c.Bool("swarm"),
			Master:             c.Bool("swarm-master"),
			Discovery:          c.String("swarm-discovery"),
			Address:            c.String("swarm-addr"),
			Host:               c.String("swarm-host"),
			Strategy:           c.String("swarm-strategy"),
			ArbitraryFlags:     c.StringSlice("swarm-opt"),
			ArbitraryJoinFlags: c.StringSlice("swarm-join-opt"),
			IsExperimental:     c.Bool("swarm-experimental"),
		},
	}

	exists, err := api.Exists(h.Name)
	if err != nil {
		return fmt.Errorf("Error checking if host exists: %s", err)
	}
	if exists {
		return mcnerror.ErrHostAlreadyExists{
			Name: h.Name,
		}
	}

	// driverOpts is the actual data we send over the wire to set the
	// driver parameters (an interface fulfilling drivers.DriverOptions,
	// concrete type rpcdriver.RpcFlags).
	mcnFlags := h.Driver.GetCreateFlags()
	driverOpts := getDriverOpts(c, mcnFlags)

	if err := h.Driver.SetConfigFromFlags(driverOpts); err != nil {
		return fmt.Errorf("Error setting machine configuration from flags provided: %s", err)
	}

	if err := api.Create(h); err != nil {
		// Wait for all the logs to reach the client
		time.Sleep(2 * time.Second)

		vBoxLog := ""
		if h.DriverName == "virtualbox" {
			vBoxLog = filepath.Join(api.GetMachinesDir(), h.Name, h.Name, "Logs", "VBox.log")
		}

		return crashreport.CrashError{
			Cause:       err,
			Command:     "Create",
			Context:     "api.performCreate",
			DriverName:  h.DriverName,
			LogFilePath: vBoxLog,
		}
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error attempting to save store: %s", err)
	}

	log.Infof("To see how to connect your Docker Client to the Docker Engine running on this virtual machine, run: %s env %s", os.Args[0], name)

	return nil
}

// The following function is needed because the CLI acrobatics that we're doing
// (with having an "outer" and "inner" function each with their own custom
// settings and flag parsing needs) are not well supported by codegangsta/cli.
//
// Instead of trying to make a convoluted series of flag parsing and relying on
// codegangsta/cli internals work well, we simply read the flags we're
// interested in from the outer function into module-level variables, and then
// use them from the "inner" function.
//
// I'm not very pleased about this, but it seems to be the only decent
// compromise without drastically modifying codegangsta/cli internals or our
// own CLI.
func flagHackLookup(flagName string) string {
	// e.g. "-d" for "--driver"
	flagPrefix := flagName[1:3]

	// TODO: Should we support -flag-name (single hyphen) syntax as well?
	for i, arg := range os.Args {
		if strings.Contains(arg, flagPrefix) {
			// format '--driver foo' or '-d foo'
			if arg == flagPrefix || arg == flagName {
				if i+1 < len(os.Args) {
					return os.Args[i+1]
				}
			}

			// format '--driver=foo' or '-d=foo'
			if strings.HasPrefix(arg, flagPrefix+"=") || strings.HasPrefix(arg, flagName+"=") {
				return strings.Split(arg, "=")[1]
			}
		}
	}

	return ""
}

func cmdCreateOuter(c CommandLine, api libmachine.API) error {
	const (
		flagLookupMachineName = "flag-lookup"
	)

	// We didn't recognize the driver name.
	driverName := flagHackLookup("--driver")
	if driverName == "" {
		//TODO: Check Environment have to include flagHackLookup function.
		driverName = os.Getenv("MACHINE_DRIVER")
		if driverName == "" {
			c.ShowHelp()
			return nil // ?
		}
	}

	// TODO: Fix hacky JSON solution
	rawDriver, err := json.Marshal(&drivers.BaseDriver{
		MachineName: flagLookupMachineName,
	})
	if err != nil {
		return fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
	}

	h, err := api.NewHost(driverName, rawDriver)
	if err != nil {
		return err
	}

	// TODO: So much flag manipulation and voodoo here, it seems to be
	// asking for trouble.
	//
	// mcnFlags is the data we get back over the wire (type mcnflag.Flag)
	// to indicate which parameters are available.
	mcnFlags := h.Driver.GetCreateFlags()

	// This bit will actually make "create" display the correct flags based
	// on the requested driver.
	cliFlags, err := convertMcnFlagsToCliFlags(mcnFlags)
	if err != nil {
		return fmt.Errorf("Error trying to convert provided driver flags to cli flags: %s", err)
	}

	for i := range c.Application().Commands {
		cmd := &c.Application().Commands[i]
		if cmd.HasName("create") {
			cmd = addDriverFlagsToCommand(cliFlags, cmd)
		}
	}

	return c.Application().Run(os.Args)
}

func getDriverOpts(c CommandLine, mcnflags []mcnflag.Flag) drivers.DriverOptions {
	// TODO: This function is pretty damn YOLO and would benefit from some
	// sanity checking around types and assertions.
	//
	// But, we need it so that we can actually send the flags for creating
	// a machine over the wire (cli.Context is a no go since there is so
	// much stuff in it).
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}

	for _, f := range mcnflags {
		driverOpts.Values[f.String()] = f.Default()

		// Hardcoded logic for boolean... :(
		if f.Default() == nil {
			driverOpts.Values[f.String()] = false
		}
	}

	for _, name := range c.FlagNames() {
		getter, ok := c.Generic(name).(flag.Getter)
		if ok {
			driverOpts.Values[name] = getter.Get()
		} else {
			// TODO: This is pretty hacky.  StringSlice is the only
			// type so far we have to worry about which is not a
			// Getter, though.
			if c.IsSet(name) {
				driverOpts.Values[name] = c.StringSlice(name)
			}
		}
	}

	return driverOpts
}

func convertMcnFlagsToCliFlags(mcnFlags []mcnflag.Flag) ([]cli.Flag, error) {
	cliFlags := []cli.Flag{}
	for _, f := range mcnFlags {
		switch t := f.(type) {
		// TODO: It seems pretty wrong to just default "nil" to this,
		// but cli.BoolFlag doesn't have a "Value" field (false is
		// always the default)
		case *mcnflag.BoolFlag:
			f := f.(*mcnflag.BoolFlag)
			cliFlags = append(cliFlags, cli.BoolFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
			})
		case *mcnflag.IntFlag:
			f := f.(*mcnflag.IntFlag)
			cliFlags = append(cliFlags, cli.IntFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
				Value:  f.Value,
			})
		case *mcnflag.StringFlag:
			f := f.(*mcnflag.StringFlag)
			cliFlags = append(cliFlags, cli.StringFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,
				Value:  f.Value,
			})
		case *mcnflag.StringSliceFlag:
			f := f.(*mcnflag.StringSliceFlag)
			cliFlags = append(cliFlags, cli.StringSliceFlag{
				Name:   f.Name,
				EnvVar: f.EnvVar,
				Usage:  f.Usage,

				//TODO: Is this used with defaults? Can we convert the literal []string to cli.StringSlice properly?
				Value: &cli.StringSlice{},
			})
		default:
			log.Warn("Flag is ", f)
			return nil, fmt.Errorf("Flag is unrecognized flag type: %T", t)
		}
	}

	return cliFlags, nil
}

func addDriverFlagsToCommand(cliFlags []cli.Flag, cmd *cli.Command) *cli.Command {
	cmd.Flags = append(SharedCreateFlags, cliFlags...)
	cmd.SkipFlagParsing = false
	cmd.Action = runCommand(cmdCreateInner)
	sort.Sort(ByFlagName(cmd.Flags))

	return cmd
}

func validateSwarmDiscovery(discovery string) error {
	if discovery == "" {
		return nil
	}

	matched, err := regexp.MatchString(`[^:]*://.*`, discovery)
	if err != nil {
		return err
	}

	if matched {
		return nil
	}

	return fmt.Errorf("Swarm Discovery URL was in the wrong format: %s", discovery)
}

func tlsPath(c CommandLine, flag string, defaultName string) string {
	path := c.GlobalString(flag)
	if path != "" {
		return path
	}

	return filepath.Join(mcndirs.GetMachineCertDir(), defaultName)
}
