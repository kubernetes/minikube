package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/crashreport"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/persist"
	"github.com/docker/machine/libmachine/ssh"
)

const (
	defaultMachineName = "default"
)

var (
	ErrHostLoad           = errors.New("All specified hosts had errors loading their configuration.")
	ErrNoDefault          = fmt.Errorf("Error: No machine name(s) specified and no %q machine exists.", defaultMachineName)
	ErrNoMachineSpecified = errors.New("Error: Expected to get one or more machine names as arguments")
	ErrExpectedOneMachine = errors.New("Error: Expected one machine name as an argument")
	ErrTooManyArguments   = errors.New("Error: Too many arguments given")

	osExit = func(code int) { os.Exit(code) }
)

// CommandLine contains all the information passed to the commands on the command line.
type CommandLine interface {
	ShowHelp()

	ShowVersion()

	Application() *cli.App

	Args() cli.Args

	IsSet(name string) bool

	Bool(name string) bool

	Int(name string) int

	String(name string) string

	StringSlice(name string) []string

	GlobalString(name string) string

	FlagNames() (names []string)

	Generic(name string) interface{}
}

type contextCommandLine struct {
	*cli.Context
}

func (c *contextCommandLine) ShowHelp() {
	cli.ShowCommandHelp(c.Context, c.Command.Name)
}

func (c *contextCommandLine) ShowVersion() {
	cli.ShowVersion(c.Context)
}

func (c *contextCommandLine) Application() *cli.App {
	return c.App
}

// targetHost returns a specific host name if one is indicated by the first CLI
// arg, or the default host name if no host is specified.
func targetHost(c CommandLine, api libmachine.API) (string, error) {
	if len(c.Args()) == 0 {
		defaultExists, err := api.Exists(defaultMachineName)
		if err != nil {
			return "", fmt.Errorf("Error checking if host %q exists: %s", defaultMachineName, err)
		}

		if defaultExists {
			return defaultMachineName, nil
		}

		return "", ErrNoDefault
	}

	return c.Args()[0], nil
}

func runAction(actionName string, c CommandLine, api libmachine.API) error {
	var (
		hostsToLoad []string
	)

	// If user did not specify a machine name explicitly, use the 'default'
	// machine if it exists.  This allows short form commands such as
	// 'docker-machine stop' for convenience.
	if len(c.Args()) == 0 {
		target, err := targetHost(c, api)
		if err != nil {
			return err
		}

		hostsToLoad = []string{target}
	} else {
		hostsToLoad = c.Args()
	}

	hosts, hostsInError := persist.LoadHosts(api, hostsToLoad)

	if len(hostsInError) > 0 {
		errs := []error{}
		for _, err := range hostsInError {
			errs = append(errs, err)
		}
		return consolidateErrs(errs)
	}

	if len(hosts) == 0 {
		return ErrHostLoad
	}

	if errs := runActionForeachMachine(actionName, hosts); len(errs) > 0 {
		return consolidateErrs(errs)
	}

	for _, h := range hosts {
		if err := api.Save(h); err != nil {
			return fmt.Errorf("Error saving host to store: %s", err)
		}
	}

	return nil
}

func runCommand(command func(commandLine CommandLine, api libmachine.API) error) func(context *cli.Context) {
	return func(context *cli.Context) {
		api := libmachine.NewClient(mcndirs.GetBaseDir(), mcndirs.GetMachineCertDir())
		defer api.Close()

		if context.GlobalBool("native-ssh") {
			api.SSHClientType = ssh.Native
		}
		api.GithubAPIToken = context.GlobalString("github-api-token")
		api.Filestore.Path = context.GlobalString("storage-path")

		// TODO (nathanleclaire): These should ultimately be accessed
		// through the libmachine client by the rest of the code and
		// not through their respective modules.  For now, however,
		// they are also being set the way that they originally were
		// set to preserve backwards compatibility.
		mcndirs.BaseDir = api.Filestore.Path
		mcnutils.GithubAPIToken = api.GithubAPIToken
		ssh.SetDefaultClient(api.SSHClientType)

		if err := command(&contextCommandLine{context}, api); err != nil {
			log.Error(err)

			if crashErr, ok := err.(crashreport.CrashError); ok {
				crashReporter := crashreport.NewCrashReporter(mcndirs.GetBaseDir(), context.GlobalString("bugsnag-api-token"))
				crashReporter.Send(crashErr)

				if _, ok := crashErr.Cause.(mcnerror.ErrDuringPreCreate); ok {
					osExit(3)
					return
				}
			}

			osExit(1)
			return
		}
	}
}

func confirmInput(msg string) (bool, error) {
	fmt.Printf("%s (y/n): ", msg)

	var resp string
	_, err := fmt.Scanln(&resp)
	if err != nil {
		return false, err
	}

	confirmed := strings.Index(strings.ToLower(resp), "y") == 0
	return confirmed, nil
}

var Commands = []cli.Command{
	{
		Name:   "active",
		Usage:  "Print which machine is active",
		Action: runCommand(cmdActive),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout, t",
				Usage: fmt.Sprintf("Timeout in seconds, default to %ds", activeDefaultTimeout),
				Value: activeDefaultTimeout,
			},
		},
	},
	{
		Name:        "config",
		Usage:       "Print the connection config for machine",
		Description: "Argument is a machine name.",
		Action:      runCommand(cmdConfig),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "swarm",
				Usage: "Display the Swarm config instead of the Docker daemon",
			},
		},
	},
	{
		Flags:           SharedCreateFlags,
		Name:            "create",
		Usage:           "Create a machine",
		Description:     fmt.Sprintf("Run '%s create --driver name' to include the create flags for that driver in the help text.", os.Args[0]),
		Action:          runCommand(cmdCreateOuter),
		SkipFlagParsing: true,
	},
	{
		Name:        "env",
		Usage:       "Display the commands to set up the environment for the Docker client",
		Description: "Argument is a machine name.",
		Action:      runCommand(cmdEnv),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "swarm",
				Usage: "Display the Swarm config instead of the Docker daemon",
			},
			cli.StringFlag{
				Name:  "shell",
				Usage: "Force environment to be configured for a specified shell: [fish, cmd, powershell], default is auto-detect",
			},
			cli.BoolFlag{
				Name:  "unset, u",
				Usage: "Unset variables instead of setting them",
			},
			cli.BoolFlag{
				Name:  "no-proxy",
				Usage: "Add machine IP to NO_PROXY environment variable",
			},
		},
	},
	{
		Name:        "inspect",
		Usage:       "Inspect information about a machine",
		Description: "Argument is a machine name.",
		Action:      runCommand(cmdInspect),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "format, f",
				Usage: "Format the output using the given go template.",
				Value: "",
			},
		},
	},
	{
		Name:        "ip",
		Usage:       "Get the IP address of a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdIP),
	},
	{
		Name:        "kill",
		Usage:       "Kill a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdKill),
	},
	{
		Name:   "ls",
		Usage:  "List machines",
		Action: runCommand(cmdLs),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet, q",
				Usage: "Enable quiet mode",
			},
			cli.StringSliceFlag{
				Name:  "filter",
				Usage: "Filter output based on conditions provided",
				Value: &cli.StringSlice{},
			},
			cli.IntFlag{
				Name:  "timeout, t",
				Usage: fmt.Sprintf("Timeout in seconds, default to %ds", lsDefaultTimeout),
				Value: lsDefaultTimeout,
			},
			cli.StringFlag{
				Name:  "format, f",
				Usage: "Pretty-print machines using a Go template",
			},
		},
	},
	{
		Name:   "provision",
		Usage:  "Re-provision existing machines",
		Action: runCommand(cmdProvision),
	},
	{
		Name:        "regenerate-certs",
		Usage:       "Regenerate TLS Certificates for a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdRegenerateCerts),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Force rebuild and do not prompt",
			},
		},
	},
	{
		Name:        "restart",
		Usage:       "Restart a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdRestart),
	},
	{
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Remove local configuration even if machine cannot be removed, also implies an automatic yes (`-y`)",
			},
			cli.BoolFlag{
				Name:  "y",
				Usage: "Assumes automatic yes to proceed with remove, without prompting further user confirmation",
			},
		},
		Name:        "rm",
		Usage:       "Remove a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdRm),
	},
	{
		Name:            "ssh",
		Usage:           "Log into or run a command on a machine with SSH.",
		Description:     "Arguments are [machine-name] [command]",
		Action:          runCommand(cmdSSH),
		SkipFlagParsing: true,
	},
	{
		Name:        "scp",
		Usage:       "Copy files between machines",
		Description: "Arguments are [machine:][path] [machine:][path].",
		Action:      runCommand(cmdScp),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "Copy files recursively (required to copy directories)",
			},
		},
	},
	{
		Name:        "start",
		Usage:       "Start a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdStart),
	},
	{
		Name:        "status",
		Usage:       "Get the status of a machine",
		Description: "Argument is a machine name.",
		Action:      runCommand(cmdStatus),
	},
	{
		Name:        "stop",
		Usage:       "Stop a machine",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdStop),
	},
	{
		Name:        "upgrade",
		Usage:       "Upgrade a machine to the latest version of Docker",
		Description: "Argument(s) are one or more machine names.",
		Action:      runCommand(cmdUpgrade),
	},
	{
		Name:        "url",
		Usage:       "Get the URL of a machine",
		Description: "Argument is a machine name.",
		Action:      runCommand(cmdURL),
	},
	{
		Name:   "version",
		Usage:  "Show the Docker Machine version or a machine docker version",
		Action: runCommand(cmdVersion),
	},
}

func printIP(h *host.Host) func() error {
	return func() error {
		ip, err := h.Driver.GetIP()
		if err != nil {
			return fmt.Errorf("Error getting IP address: %s", err)
		}

		fmt.Println(ip)

		return nil
	}
}

// machineCommand maps the command name to the corresponding machine command.
// We run commands concurrently and communicate back an error if there was one.
func machineCommand(actionName string, host *host.Host, errorChan chan<- error) {
	// TODO: These actions should have their own type.
	commands := map[string](func() error){
		"configureAuth": host.ConfigureAuth,
		"start":         host.Start,
		"stop":          host.Stop,
		"restart":       host.Restart,
		"kill":          host.Kill,
		"upgrade":       host.Upgrade,
		"ip":            printIP(host),
		"provision":     host.Provision,
	}

	log.Debugf("command=%s machine=%s", actionName, host.Name)

	errorChan <- commands[actionName]()
}

// runActionForeachMachine will run the command across multiple machines
func runActionForeachMachine(actionName string, machines []*host.Host) []error {
	var (
		numConcurrentActions = 0
		errorChan            = make(chan error)
		errs                 = []error{}
	)

	for _, machine := range machines {
		numConcurrentActions++
		go machineCommand(actionName, machine, errorChan)
	}

	// TODO: We should probably only do 5-10 of these
	// at a time, since otherwise cloud providers might
	// rate limit us.
	for i := 0; i < numConcurrentActions; i++ {
		if err := <-errorChan; err != nil {
			errs = append(errs, err)
		}
	}

	close(errorChan)

	return errs
}

func consolidateErrs(errs []error) error {
	finalErr := ""
	for _, err := range errs {
		finalErr = fmt.Sprintf("%s\n%s", finalErr, err)
	}

	return errors.New(strings.TrimSpace(finalErr))
}
