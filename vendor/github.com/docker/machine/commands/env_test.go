package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/check"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

type FakeConnChecker struct {
	DockerHost  string
	AuthOptions *auth.Options
	Err         error
}

func (fcc *FakeConnChecker) Check(_ *host.Host, _ bool) (string, *auth.Options, error) {
	return fcc.DockerHost, fcc.AuthOptions, fcc.Err
}

type SimpleUsageHintGenerator struct {
	Hint string
}

func (suhg *SimpleUsageHintGenerator) GenerateUsageHint(_ string, _ []string) string {
	return suhg.Hint
}

func TestHints(t *testing.T) {
	var tests = []struct {
		userShell     string
		commandLine   []string
		expectedHints string
	}{
		{"", []string{"machine", "env", "default"}, "# Run this command to configure your shell: \n# eval $(machine env default)\n"},
		{"", []string{"machine", "env", "--no-proxy", "default"}, "# Run this command to configure your shell: \n# eval $(machine env --no-proxy default)\n"},
		{"", []string{"machine", "env", "--swarm", "default"}, "# Run this command to configure your shell: \n# eval $(machine env --swarm default)\n"},
		{"", []string{"machine", "env", "--no-proxy", "--swarm", "default"}, "# Run this command to configure your shell: \n# eval $(machine env --no-proxy --swarm default)\n"},
		{"", []string{"machine", "env", "--unset"}, "# Run this command to configure your shell: \n# eval $(machine env --unset)\n"},
		{"", []string{`C:\\Program Files\docker-machine.exe`, "env", "default"}, "# Run this command to configure your shell: \n# eval $(\"C:\\\\Program Files\\docker-machine.exe\" env default)\n"},
		{"", []string{`C:\\Me\docker-machine.exe`, "env", "default"}, "# Run this command to configure your shell: \n# eval $(\"C:\\\\Me\\docker-machine.exe\" env default)\n"},

		{"fish", []string{"./machine", "env", "--shell=fish", "default"}, "# Run this command to configure your shell: \n# eval (./machine env --shell=fish default)\n"},
		{"fish", []string{"./machine", "env", "--shell=fish", "--no-proxy", "default"}, "# Run this command to configure your shell: \n# eval (./machine env --shell=fish --no-proxy default)\n"},
		{"fish", []string{"./machine", "env", "--shell=fish", "--swarm", "default"}, "# Run this command to configure your shell: \n# eval (./machine env --shell=fish --swarm default)\n"},
		{"fish", []string{"./machine", "env", "--shell=fish", "--no-proxy", "--swarm", "default"}, "# Run this command to configure your shell: \n# eval (./machine env --shell=fish --no-proxy --swarm default)\n"},
		{"fish", []string{"./machine", "env", "--shell=fish", "--unset"}, "# Run this command to configure your shell: \n# eval (./machine env --shell=fish --unset)\n"},

		{"powershell", []string{"./machine", "env", "--shell=powershell", "default"}, "# Run this command to configure your shell: \n# & ./machine env --shell=powershell default | Invoke-Expression\n"},
		{"powershell", []string{"./machine", "env", "--shell=powershell", "--no-proxy", "default"}, "# Run this command to configure your shell: \n# & ./machine env --shell=powershell --no-proxy default | Invoke-Expression\n"},
		{"powershell", []string{"./machine", "env", "--shell=powershell", "--swarm", "default"}, "# Run this command to configure your shell: \n# & ./machine env --shell=powershell --swarm default | Invoke-Expression\n"},
		{"powershell", []string{"./machine", "env", "--shell=powershell", "--no-proxy", "--swarm", "default"}, "# Run this command to configure your shell: \n# & ./machine env --shell=powershell --no-proxy --swarm default | Invoke-Expression\n"},
		{"powershell", []string{"./machine", "env", "--shell=powershell", "--unset"}, "# Run this command to configure your shell: \n# & ./machine env --shell=powershell --unset | Invoke-Expression\n"},
		{"powershell", []string{"./machine", "env", "--shell=powershell", "--unset"}, "# Run this command to configure your shell: \n# & ./machine env --shell=powershell --unset | Invoke-Expression\n"},
		{"powershell", []string{`C:\\Program Files\docker-machine.exe`, "env", "--shell=powershell", "default"}, "# Run this command to configure your shell: \n# & \"C:\\\\Program Files\\docker-machine.exe\" env --shell=powershell default | Invoke-Expression\n"},
		{"powershell", []string{`C:\\Me\docker-machine.exe`, "env", "--shell=powershell", "default"}, "# Run this command to configure your shell: \n# & \"C:\\\\Me\\docker-machine.exe\" env --shell=powershell default | Invoke-Expression\n"},

		{"cmd", []string{"./machine", "env", "--shell=cmd", "default"}, "REM Run this command to configure your shell: \nREM \t@FOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd default') DO @%i\n"},
		{"cmd", []string{"./machine", "env", "--shell=cmd", "--no-proxy", "default"}, "REM Run this command to configure your shell: \nREM \t@FOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd --no-proxy default') DO @%i\n"},
		{"cmd", []string{"./machine", "env", "--shell=cmd", "--swarm", "default"}, "REM Run this command to configure your shell: \nREM \t@FOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd --swarm default') DO @%i\n"},
		{"cmd", []string{"./machine", "env", "--shell=cmd", "--no-proxy", "--swarm", "default"}, "REM Run this command to configure your shell: \nREM \t@FOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd --no-proxy --swarm default') DO @%i\n"},
		{"cmd", []string{"./machine", "env", "--shell=cmd", "--unset"}, "REM Run this command to configure your shell: \nREM \t@FOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd --unset') DO @%i\n"},
		{"cmd", []string{`C:\\Program Files\docker-machine.exe`, "env", "--shell=cmd", "default"}, "REM Run this command to configure your shell: \nREM \t@FOR /f \"tokens=*\" %i IN ('\"C:\\\\Program Files\\docker-machine.exe\" env --shell=cmd default') DO @%i\n"},
		{"cmd", []string{`C:\\Me\docker-machine.exe`, "env", "--shell=cmd", "default"}, "REM Run this command to configure your shell: \nREM \t@FOR /f \"tokens=*\" %i IN ('\"C:\\\\Me\\docker-machine.exe\" env --shell=cmd default') DO @%i\n"},

		{"emacs", []string{"./machine", "env", "--shell=emacs", "default"}, ";; Run this command to configure your shell: \n;; (with-temp-buffer (shell-command \"./machine env --shell=emacs default\" (current-buffer)) (eval-buffer))\n"},
		{"emacs", []string{"./machine", "env", "--shell=emacs", "--no-proxy", "default"}, ";; Run this command to configure your shell: \n;; (with-temp-buffer (shell-command \"./machine env --shell=emacs --no-proxy default\" (current-buffer)) (eval-buffer))\n"},
		{"emacs", []string{"./machine", "env", "--shell=emacs", "--swarm", "default"}, ";; Run this command to configure your shell: \n;; (with-temp-buffer (shell-command \"./machine env --shell=emacs --swarm default\" (current-buffer)) (eval-buffer))\n"},
		{"emacs", []string{"./machine", "env", "--shell=emacs", "--no-proxy", "--swarm", "default"}, ";; Run this command to configure your shell: \n;; (with-temp-buffer (shell-command \"./machine env --shell=emacs --no-proxy --swarm default\" (current-buffer)) (eval-buffer))\n"},
		{"emacs", []string{"./machine", "env", "--shell=emacs", "--unset"}, ";; Run this command to configure your shell: \n;; (with-temp-buffer (shell-command \"./machine env --shell=emacs --unset\" (current-buffer)) (eval-buffer))\n"},
	}

	for _, test := range tests {
		hints := defaultUsageHinter.GenerateUsageHint(test.userShell, test.commandLine)
		assert.Equal(t, test.expectedHints, hints)
	}
}

func revertUsageHinter(uhg UsageHintGenerator) {
	defaultUsageHinter = uhg
}

func TestShellCfgSet(t *testing.T) {
	const (
		usageHint = "This is a usage hint"
	)

	// TODO: This should be embedded in some kind of wrapper struct for all
	// these `env` operations.
	defer revertUsageHinter(defaultUsageHinter)
	defaultUsageHinter = &SimpleUsageHintGenerator{usageHint}

	var tests = []struct {
		description      string
		commandLine      CommandLine
		api              libmachine.API
		connChecker      check.ConnChecker
		noProxyVar       string
		noProxyValue     string
		expectedShellCfg *ShellConfig
		expectedErr      error
	}{
		{
			description: "no host name specified",
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{},
			},
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: nil,
			},
			expectedShellCfg: nil,
			expectedErr:      ErrNoDefault,
		},
		{
			description: "bash shell set happy path without any flags set",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"quux"},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "bash",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "quux",
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "export ",
				Delimiter:       "=\"",
				Suffix:          "\"\n",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), "quux"),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				MachineName:     "quux",
			},
			expectedErr: nil,
		},
		{
			description: "bash shell set happy path with 'default' vm",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "bash",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: defaultMachineName,
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "export ",
				Delimiter:       "=\"",
				Suffix:          "\"\n",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), defaultMachineName),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				MachineName:     defaultMachineName,
			},
			expectedErr: nil,
		},
		{
			description: "fish shell set happy path",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"quux"},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "fish",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "quux",
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "set -gx ",
				Suffix:          "\";\n",
				Delimiter:       " \"",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), "quux"),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				MachineName:     "quux",
			},
			expectedErr: nil,
		},
		{
			description: "powershell set happy path",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"quux"},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "powershell",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "quux",
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "$Env:",
				Suffix:          "\"\n",
				Delimiter:       " = \"",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), "quux"),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				MachineName:     "quux",
			},
			expectedErr: nil,
		},
		{
			description: "emacs set happy path",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"quux"},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "emacs",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "quux",
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "(setenv \"",
				Suffix:          "\")\n",
				Delimiter:       "\" \"",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), "quux"),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				MachineName:     "quux",
			},
			expectedErr: nil,
		},
		{
			description: "cmd.exe happy path",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"quux"},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "cmd",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "quux",
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "SET ",
				Suffix:          "\n",
				Delimiter:       "=",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), "quux"),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				MachineName:     "quux",
			},
			expectedErr: nil,
		},
		{
			description: "bash shell set happy path with --no-proxy flag; no existing environment variable set",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"quux"},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "bash",
						"swarm":    false,
						"no-proxy": true,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "quux",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
							MockIP:    "1.2.3.4",
						},
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "export ",
				Delimiter:       "=\"",
				Suffix:          "\"\n",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), "quux"),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				NoProxyVar:      "NO_PROXY",
				NoProxyValue:    "1.2.3.4", // From FakeDriver
				MachineName:     "quux",
			},
			noProxyVar:   "NO_PROXY",
			noProxyValue: "",
			expectedErr:  nil,
		},
		{
			description: "bash shell set happy path with --no-proxy flag; existing environment variable _is_ set",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"quux"},
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "bash",
						"swarm":    false,
						"no-proxy": true,
					},
				},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "quux",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
							MockIP:    "1.2.3.4",
						},
					},
				},
			},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:          "export ",
				Delimiter:       "=\"",
				Suffix:          "\"\n",
				DockerCertPath:  filepath.Join(mcndirs.GetMachineDir(), "quux"),
				DockerHost:      "tcp://1.2.3.4:2376",
				DockerTLSVerify: "1",
				UsageHint:       usageHint,
				NoProxyVar:      "no_proxy",
				NoProxyValue:    "192.168.59.1,1.2.3.4", // From FakeDriver
				MachineName:     "quux",
			},
			noProxyVar:   "no_proxy",
			noProxyValue: "192.168.59.1",
			expectedErr:  nil,
		},
	}

	for _, test := range tests {
		// TODO: Ideally this should not hit the environment at all but
		// rather should go through an interface.
		os.Setenv(test.noProxyVar, test.noProxyValue)

		t.Log(test.description)

		check.DefaultConnChecker = test.connChecker
		shellCfg, err := shellCfgSet(test.commandLine, test.api)
		assert.Equal(t, test.expectedShellCfg, shellCfg)
		assert.Equal(t, test.expectedErr, err)

		os.Unsetenv(test.noProxyVar)
	}
}

func TestShellCfgUnset(t *testing.T) {
	const (
		usageHint = "This is the unset usage hint"
	)

	defer revertUsageHinter(defaultUsageHinter)
	defaultUsageHinter = &SimpleUsageHintGenerator{usageHint}

	var tests = []struct {
		description      string
		commandLine      CommandLine
		api              libmachine.API
		connChecker      check.ConnChecker
		noProxyVar       string
		noProxyValue     string
		expectedShellCfg *ShellConfig
		expectedErr      error
	}{
		{
			description: "more than expected args passed in",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"foo", "bar"},
			},
			expectedShellCfg: nil,
			expectedErr:      errImproperUnsetEnvArgs,
		},
		{
			description: "bash shell unset happy path without any flags set",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: nil,
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "bash",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:    "unset ",
				Suffix:    "\n",
				Delimiter: "",
				UsageHint: usageHint,
			},
			expectedErr: nil,
		},
		{
			description: "fish shell unset happy path",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: nil,
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "fish",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:    "set -e ",
				Suffix:    ";\n",
				Delimiter: "",
				UsageHint: usageHint,
			},
			expectedErr: nil,
		},
		{
			description: "powershell unset happy path",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: nil,
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "powershell",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:    `Remove-Item Env:\\`,
				Suffix:    "\n",
				Delimiter: "",
				UsageHint: usageHint,
			},
			expectedErr: nil,
		},
		{
			description: "cmd.exe unset happy path",
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: nil,
				LocalFlags: &commandstest.FakeFlagger{
					Data: map[string]interface{}{
						"shell":    "cmd",
						"swarm":    false,
						"no-proxy": false,
					},
				},
			},
			api: &libmachinetest.FakeAPI{},
			connChecker: &FakeConnChecker{
				DockerHost:  "tcp://1.2.3.4:2376",
				AuthOptions: nil,
				Err:         nil,
			},
			expectedShellCfg: &ShellConfig{
				Prefix:    "SET ",
				Suffix:    "\n",
				Delimiter: "=",
				UsageHint: usageHint,
			},
			expectedErr: nil,
		},
		// TODO: There is kind of a funny bug (feature?) I discovered
		// reasoning about unset() where if there was a NO_PROXY value
		// set _before_ the original docker-machine env, it won't be
		// restored (NO_PROXY won't be unset at all, it will stay the
		// same).  We should define expected behavior in this case.
	}

	for _, test := range tests {
		os.Setenv(test.noProxyVar, test.noProxyValue)

		t.Log(test.description)

		check.DefaultConnChecker = test.connChecker
		shellCfg, err := shellCfgUnset(test.commandLine, test.api)
		assert.Equal(t, test.expectedShellCfg, shellCfg)
		assert.Equal(t, test.expectedErr, err)

		os.Setenv(test.noProxyVar, "")
	}
}
