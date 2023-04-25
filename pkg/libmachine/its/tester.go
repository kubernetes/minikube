package its

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"io/ioutil"

	"os"

	"fmt"

	"path/filepath"
	"runtime"
)

var (
	regexpCommandLine = regexp.MustCompile("('[^']*')|(\\S+)")
)

type IntegrationTest interface {
	RequireDriver(driverName string)

	SkipDriver(driverName string)

	SkipDrivers(driverNames ...string)

	ForceDriver(driverName string)

	Run(description string, action func())

	Cmd(commandLine string) IntegrationTest

	Machine(commandLine string) IntegrationTest

	DriverName() string

	Should() Assertions

	TearDown()
}

type Assertions interface {
	Succeed(messages ...string) Assertions

	Fail(errorMessages ...string) Assertions

	ContainLines(count int) Assertions

	ContainLine(index int, text string) Assertions

	MatchLine(index int, template string) Assertions

	EqualLine(index int, text string) Assertions
}

func NewTest(t *testing.T) IntegrationTest {
	storagePath, _ := ioutil.TempDir("", "docker")

	return &dockerMachineTest{
		t:           t,
		storagePath: storagePath,
	}
}

type dockerMachineTest struct {
	t                   *testing.T
	storagePath         string
	dockerMachineBinary string
	description         string
	skip                bool
	rawOutput           string
	lines               []string
	err                 error
	fatal               bool
	failed              bool
}

func (dmt *dockerMachineTest) RequireDriver(driverName string) {
	dmt.skipIf(dmt.DriverName() != driverName)
}

func (dmt *dockerMachineTest) SkipDriver(driverName string) {
	dmt.skipIf(dmt.DriverName() == driverName)
}

func (dmt *dockerMachineTest) SkipDrivers(driverNames ...string) {
	for _, driverName := range driverNames {
		dmt.skipIf(dmt.DriverName() == driverName)
	}
}

func (dmt *dockerMachineTest) ForceDriver(driverName string) {
	os.Setenv("DRIVER", driverName)
}

func (dmt *dockerMachineTest) skipIf(condition bool) {
	if condition {
		dmt.skip = true
	}
}

func (dmt *dockerMachineTest) Run(description string, action func()) {
	dmt.description = description
	dmt.rawOutput = ""
	dmt.lines = nil
	dmt.err = nil
	dmt.failed = false

	if dmt.skip {
		fmt.Printf("%s %s\n", yellow("[SKIP]"), description)
	} else {
		fmt.Printf("%s %s", yellow("[..]"), description)
		action()

		if dmt.fatal || dmt.failed {
			fmt.Printf("\r%s %s\n", red("[KO]"), description)
		} else {
			fmt.Printf("\r%s %s\n", green("[OK]"), description)
		}
	}
}

func red(message string) string {
	if runtime.GOOS == "windows" {
		return message
	}
	return "\033[1;31m" + message + "\033[0m"
}

func green(message string) string {
	if runtime.GOOS == "windows" {
		return message
	}
	return "\033[1;32m" + message + "\033[0m"
}

func yellow(message string) string {
	if runtime.GOOS == "windows" {
		return message
	}
	return "\033[1;33m" + message + "\033[0m"
}

func (dmt *dockerMachineTest) DriverName() string {
	driver := os.Getenv("DRIVER")
	if driver != "" {
		return driver
	}

	return "virtualbox"
}

func (dmt *dockerMachineTest) Should() Assertions {
	return dmt
}

func (dmt *dockerMachineTest) testedBinary() string {
	if dmt.dockerMachineBinary != "" {
		return dmt.dockerMachineBinary
	}

	var binary string
	if runtime.GOOS == "windows" {
		binary = "docker-machine.exe"
	} else {
		binary = "docker-machine"
	}

	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	for dir != "/" {
		path := filepath.Join(dir, "bin", binary)

		_, err := os.Stat(path)
		if err == nil {
			dmt.dockerMachineBinary = path
			return path
		}

		dir = filepath.Dir(dir)
	}

	if !dmt.fatal {
		dmt.fatal = true
		dmt.t.Errorf("Binary not found: %s", binary)
	}

	return ""
}

func (dmt *dockerMachineTest) Cmd(commandLine string) IntegrationTest {
	if dmt.fatal {
		return dmt
	}

	commandLine = dmt.replaceDriver(commandLine)
	commandLine = dmt.replaceMachinePath(commandLine)

	return dmt.cmd("bash", "-c", commandLine)
}

func (dmt *dockerMachineTest) Machine(commandLine string) IntegrationTest {
	if dmt.fatal {
		return dmt
	}

	commandLine = dmt.replaceDriver(commandLine)

	return dmt.cmd(dmt.testedBinary(), parseFields(commandLine)...)
}

func (dmt *dockerMachineTest) cmd(command string, args ...string) IntegrationTest {
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), "MACHINE_STORAGE_PATH="+dmt.storagePath)

	combinedOutput, err := cmd.CombinedOutput()

	dmt.rawOutput = string(combinedOutput)
	dmt.lines = strings.Split(strings.TrimSpace(dmt.rawOutput), "\n")
	dmt.err = err

	return dmt
}

func (dmt *dockerMachineTest) replaceMachinePath(commandLine string) string {
	return strings.Replace(commandLine, "machine", dmt.testedBinary(), -1)
}

func (dmt *dockerMachineTest) replaceDriver(commandLine string) string {
	return strings.Replace(commandLine, "$DRIVER", dmt.DriverName(), -1)
}

func parseFields(commandLine string) []string {
	fields := regexpCommandLine.FindAllString(commandLine, -1)

	for i := range fields {
		if len(fields[i]) > 2 && strings.HasPrefix(fields[i], "'") && strings.HasSuffix(fields[i], "'") {
			fields[i] = fields[i][1 : len(fields[i])-1]
		}
	}

	return fields
}

func (dmt *dockerMachineTest) TearDown() {
	machines := filepath.Join(dmt.storagePath, "machines")

	dirs, _ := ioutil.ReadDir(machines)
	for _, dir := range dirs {
		dmt.Cmd("machine rm -f " + dir.Name())
	}

	os.RemoveAll(dmt.storagePath)
}

func (dmt *dockerMachineTest) ContainLines(count int) Assertions {
	if dmt.fatal {
		return dmt
	}

	if count != len(dmt.lines) {
		return dmt.failExpected("%d lines but got %d\n%s", count, len(dmt.lines), dmt.rawOutput)
	}

	return dmt
}

func (dmt *dockerMachineTest) ContainLine(index int, text string) Assertions {
	if dmt.fatal {
		return dmt
	}

	if index >= len(dmt.lines) {
		return dmt.failExpected("at least %d lines\nGot %d", index+1, len(dmt.lines))
	}

	if !strings.Contains(dmt.lines[index], text) {
		return dmt.failExpected("line %d to contain '%s'\nGot '%s'", index, text, dmt.lines[index])
	}

	return dmt
}

func (dmt *dockerMachineTest) MatchLine(index int, template string) Assertions {
	if dmt.fatal {
		return dmt
	}

	if index >= len(dmt.lines) {
		return dmt.failExpected("at least %d lines\nGot %d", index+1, len(dmt.lines))
	}

	if !regexp.MustCompile(template).MatchString(dmt.lines[index]) {
		return dmt.failExpected("line %d to match '%s'\nGot '%s'", index, template, dmt.lines[index])
	}

	return dmt
}

func (dmt *dockerMachineTest) EqualLine(index int, text string) Assertions {
	if dmt.fatal {
		return dmt
	}

	if index >= len(dmt.lines) {
		return dmt.failExpected("at least %d lines\nGot %d", index+1, len(dmt.lines))
	}

	if text != dmt.lines[index] {
		return dmt.failExpected("line %d to be '%s'\nGot '%s'", index, text, dmt.lines[index])
	}

	return dmt
}

func (dmt *dockerMachineTest) Succeed(messages ...string) Assertions {
	if dmt.fatal {
		return dmt
	}

	if dmt.err != nil {
		return dmt.failExpected("to succeed\nFailed with %s\n%s", dmt.err, dmt.rawOutput)
	}

	for _, message := range messages {
		if !strings.Contains(dmt.rawOutput, message) {
			return dmt.failExpected("output to contain '%s'\nGot '%s'", message, dmt.rawOutput)
		}
	}

	return dmt
}

func (dmt *dockerMachineTest) Fail(errorMessages ...string) Assertions {
	if dmt.fatal {
		return dmt
	}

	if dmt.err == nil {
		return dmt.failExpected("to fail\nGot success\n%s", dmt.rawOutput)
	}

	for _, message := range errorMessages {
		if !strings.Contains(dmt.rawOutput, message) {
			return dmt.failExpected("output to contain '%s'\nGot '%s'", message, dmt.rawOutput)
		}
	}

	return dmt
}

func (dmt *dockerMachineTest) failExpected(message string, args ...interface{}) Assertions {
	allArgs := append([]interface{}{dmt.description}, args...)

	dmt.failed = true
	dmt.t.Errorf("%s\nExpected "+message, allArgs...)

	return dmt
}
