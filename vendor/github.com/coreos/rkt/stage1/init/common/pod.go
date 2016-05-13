// Copyright 2014 The rkt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//+build linux

package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/rkt/pkg/acl"
	stage1commontypes "github.com/coreos/rkt/stage1/common/types"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/coreos/go-systemd/unit"
	"github.com/hashicorp/errwrap"

	"github.com/coreos/rkt/common"
	"github.com/coreos/rkt/common/cgroup"
	"github.com/coreos/rkt/pkg/uid"
)

const (
	// FlavorFile names the file storing the pod's flavor
	FlavorFile    = "flavor"
	sharedVolPerm = os.FileMode(0755)
)

var (
	defaultEnv = map[string]string{
		"PATH":    "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"SHELL":   "/bin/sh",
		"USER":    "root",
		"LOGNAME": "root",
		"HOME":    "/root",
	}
)

// execEscape uses Golang's string quoting for ", \, \n, and regex for special cases
func execEscape(i int, str string) string {
	escapeMap := map[string]string{
		`'`: `\`,
	}

	if i > 0 { // These are escaped only after the first argument
		escapeMap[`$`] = `$`
	}

	escArg := fmt.Sprintf("%q", str)
	for k := range escapeMap {
		reStr := `([` + regexp.QuoteMeta(k) + `])`
		re := regexp.MustCompile(reStr)
		escArg = re.ReplaceAllStringFunc(escArg, func(s string) string {
			escaped := escapeMap[s] + s
			return escaped
		})
	}
	return escArg
}

// quoteExec returns an array of quoted strings appropriate for systemd execStart usage
func quoteExec(exec []string) string {
	if len(exec) == 0 {
		// existing callers prefix {"/appexec", "/app/root", "/work/dir", "/env/file"} so this shouldn't occur.
		panic("empty exec")
	}

	var qexec []string
	for i, arg := range exec {
		escArg := execEscape(i, arg)
		qexec = append(qexec, escArg)
	}
	return strings.Join(qexec, " ")
}

// WriteDefaultTarget writes the default.target unit file
// which is responsible for bringing up the applications
func WriteDefaultTarget(p *stage1commontypes.Pod) error {
	opts := []*unit.UnitOption{
		unit.NewUnitOption("Unit", "Description", "rkt apps target"),
		unit.NewUnitOption("Unit", "DefaultDependencies", "false"),
	}

	for i := range p.Manifest.Apps {
		ra := &p.Manifest.Apps[i]
		serviceName := ServiceUnitName(ra.Name)
		opts = append(opts, unit.NewUnitOption("Unit", "After", serviceName))
		opts = append(opts, unit.NewUnitOption("Unit", "Wants", serviceName))
	}

	unitsPath := filepath.Join(common.Stage1RootfsPath(p.Root), UnitsDir)
	file, err := os.OpenFile(filepath.Join(unitsPath, "default.target"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = io.Copy(file, unit.Serialize(opts)); err != nil {
		return err
	}

	return nil
}

// WritePrepareAppTemplate writes service unit files for preparing the pod's applications
func WritePrepareAppTemplate(p *stage1commontypes.Pod) error {
	opts := []*unit.UnitOption{
		unit.NewUnitOption("Unit", "Description", "Prepare minimum environment for chrooted applications"),
		unit.NewUnitOption("Unit", "DefaultDependencies", "false"),
		unit.NewUnitOption("Unit", "OnFailureJobMode", "fail"),
		unit.NewUnitOption("Unit", "Requires", "systemd-journald.service"),
		unit.NewUnitOption("Unit", "After", "systemd-journald.service"),
		unit.NewUnitOption("Service", "Type", "oneshot"),
		unit.NewUnitOption("Service", "Restart", "no"),
		unit.NewUnitOption("Service", "ExecStart", "/prepare-app %I"),
		unit.NewUnitOption("Service", "User", "0"),
		unit.NewUnitOption("Service", "Group", "0"),
		unit.NewUnitOption("Service", "CapabilityBoundingSet", "CAP_SYS_ADMIN CAP_DAC_OVERRIDE"),
	}

	unitsPath := filepath.Join(common.Stage1RootfsPath(p.Root), UnitsDir)
	file, err := os.OpenFile(filepath.Join(unitsPath, "prepare-app@.service"), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errwrap.Wrap(errors.New("failed to create service unit file"), err)
	}
	defer file.Close()

	if _, err = io.Copy(file, unit.Serialize(opts)); err != nil {
		return errwrap.Wrap(errors.New("failed to write service unit file"), err)
	}

	return nil
}

func writeAppReaper(p *stage1commontypes.Pod, appName string) error {
	opts := []*unit.UnitOption{
		unit.NewUnitOption("Unit", "Description", fmt.Sprintf("%s Reaper", appName)),
		unit.NewUnitOption("Unit", "DefaultDependencies", "false"),
		unit.NewUnitOption("Unit", "StopWhenUnneeded", "yes"),
		unit.NewUnitOption("Unit", "Wants", "shutdown.service"),
		unit.NewUnitOption("Unit", "After", "shutdown.service"),
		unit.NewUnitOption("Unit", "Conflicts", "exit.target"),
		unit.NewUnitOption("Unit", "Conflicts", "halt.target"),
		unit.NewUnitOption("Unit", "Conflicts", "poweroff.target"),
		unit.NewUnitOption("Service", "RemainAfterExit", "yes"),
		unit.NewUnitOption("Service", "ExecStop", fmt.Sprintf("/reaper.sh %s", appName)),
	}

	unitsPath := filepath.Join(common.Stage1RootfsPath(p.Root), UnitsDir)
	file, err := os.OpenFile(filepath.Join(unitsPath, fmt.Sprintf("reaper-%s.service", appName)), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errwrap.Wrap(errors.New("failed to create service unit file"), err)
	}
	defer file.Close()

	if _, err = io.Copy(file, unit.Serialize(opts)); err != nil {
		return errwrap.Wrap(errors.New("failed to write service unit file"), err)
	}

	return nil
}

// SetJournalPermissions sets ACLs and permissions so the rkt group can access
// the pod's logs
func SetJournalPermissions(p *stage1commontypes.Pod) error {
	s1 := common.Stage1ImagePath(p.Root)

	rktgid, err := common.LookupGid(common.RktGroup)
	if err != nil {
		return fmt.Errorf("group %q not found", common.RktGroup)
	}

	journalPath := filepath.Join(s1, "rootfs", "var", "log", "journal")
	if err := os.MkdirAll(journalPath, os.FileMode(0755)); err != nil {
		return errwrap.Wrap(errors.New("error creating journal dir"), err)
	}

	a, err := acl.InitACL()
	if err != nil {
		return err
	}
	defer a.Free()

	if err := a.ParseACL(fmt.Sprintf("g:%d:r-x,m:r-x", rktgid)); err != nil {
		return errwrap.Wrap(errors.New("error parsing ACL string"), err)
	}

	if err := a.AddBaseEntries(journalPath); err != nil {
		return errwrap.Wrap(errors.New("error adding base ACL entries"), err)
	}

	if err := a.Valid(); err != nil {
		return err
	}

	if err := a.SetFileACLDefault(journalPath); err != nil {
		return errwrap.Wrap(fmt.Errorf("error setting default ACLs on %q", journalPath), err)
	}

	return nil
}

func generateGidArg(gid int, supplGid []int) string {
	arg := []string{strconv.Itoa(gid)}
	for _, sg := range supplGid {
		arg = append(arg, strconv.Itoa(sg))
	}
	return strings.Join(arg, ",")
}

// appToSystemd transforms the provided RuntimeApp+ImageManifest into systemd units
func appToSystemd(p *stage1commontypes.Pod, ra *schema.RuntimeApp, interactive bool, flavor string, privateUsers string) error {
	app := ra.App
	appName := ra.Name
	image, ok := p.Images[appName.String()]
	if !ok {
		// This is impossible as we have updated the map in LoadPod().
		panic(fmt.Sprintf("No images for app %q", ra.Name.String()))
	}
	imgName := image.Name

	if len(app.Exec) == 0 {
		return fmt.Errorf(`image %q has an empty "exec" (try --exec=BINARY)`, imgName)
	}

	workDir := "/"
	if app.WorkingDirectory != "" {
		workDir = app.WorkingDirectory
	}

	env := app.Environment

	env.Set("AC_APP_NAME", appName.String())
	if p.MetadataServiceURL != "" {
		env.Set("AC_METADATA_URL", p.MetadataServiceURL)
	}

	if err := writeEnvFile(p, env, appName, privateUsers); err != nil {
		return errwrap.Wrap(errors.New("unable to write environment file"), err)
	}

	// This is a partial implementation for app.User and app.Group:
	// For now, only numeric ids (and the string "root") are supported.
	var uid, gid int
	var err error
	if app.User == "root" {
		uid = 0
	} else {
		uid, err = strconv.Atoi(app.User)
		if err != nil {
			return fmt.Errorf("non-numerical user id not supported yet")
		}
	}
	if app.Group == "root" {
		gid = 0
	} else {
		gid, err = strconv.Atoi(app.Group)
		if err != nil {
			return fmt.Errorf("non-numerical group id not supported yet")
		}
	}

	execWrap := []string{"/appexec", common.RelAppRootfsPath(appName), workDir, RelEnvFilePath(appName), strconv.Itoa(uid), generateGidArg(gid, app.SupplementaryGIDs)}
	execStart := quoteExec(append(execWrap, app.Exec...))
	opts := []*unit.UnitOption{
		unit.NewUnitOption("Unit", "Description", fmt.Sprintf("Application=%v Image=%v", appName, imgName)),
		unit.NewUnitOption("Unit", "DefaultDependencies", "false"),
		unit.NewUnitOption("Unit", "Wants", fmt.Sprintf("reaper-%s.service", appName)),
		unit.NewUnitOption("Service", "Restart", "no"),
		unit.NewUnitOption("Service", "ExecStart", execStart),
		unit.NewUnitOption("Service", "User", "0"),
		unit.NewUnitOption("Service", "Group", "0"),
	}

	if interactive {
		opts = append(opts, unit.NewUnitOption("Service", "StandardInput", "tty"))
		opts = append(opts, unit.NewUnitOption("Service", "StandardOutput", "tty"))
		opts = append(opts, unit.NewUnitOption("Service", "StandardError", "tty"))
	} else {
		opts = append(opts, unit.NewUnitOption("Service", "StandardOutput", "journal+console"))
		opts = append(opts, unit.NewUnitOption("Service", "StandardError", "journal+console"))
		opts = append(opts, unit.NewUnitOption("Service", "SyslogIdentifier", filepath.Base(app.Exec[0])))
	}

	// When an app fails, we shut down the pod
	opts = append(opts, unit.NewUnitOption("Unit", "OnFailure", "halt.target"))

	for _, eh := range app.EventHandlers {
		var typ string
		switch eh.Name {
		case "pre-start":
			typ = "ExecStartPre"
		case "post-stop":
			typ = "ExecStopPost"
		default:
			return fmt.Errorf("unrecognized eventHandler: %v", eh.Name)
		}
		exec := quoteExec(append(execWrap, eh.Exec...))
		opts = append(opts, unit.NewUnitOption("Service", typ, exec))
	}

	// Some pre-start jobs take a long time, set the timeout to 0
	opts = append(opts, unit.NewUnitOption("Service", "TimeoutStartSec", "0"))

	var saPorts []types.Port
	for _, p := range app.Ports {
		if p.SocketActivated {
			saPorts = append(saPorts, p)
		}
	}

	for _, i := range app.Isolators {
		switch v := i.Value().(type) {
		case *types.ResourceMemory:
			opts, err = cgroup.MaybeAddIsolator(opts, "memory", v.Limit())
			if err != nil {
				return err
			}
		case *types.ResourceCPU:
			opts, err = cgroup.MaybeAddIsolator(opts, "cpu", v.Limit())
			if err != nil {
				return err
			}
		}
	}

	if len(saPorts) > 0 {
		sockopts := []*unit.UnitOption{
			unit.NewUnitOption("Unit", "Description", fmt.Sprintf("Application=%v Image=%v %s", appName, imgName, "socket-activated ports")),
			unit.NewUnitOption("Unit", "DefaultDependencies", "false"),
			unit.NewUnitOption("Socket", "BindIPv6Only", "both"),
			unit.NewUnitOption("Socket", "Service", ServiceUnitName(appName)),
		}

		for _, sap := range saPorts {
			var proto string
			switch sap.Protocol {
			case "tcp":
				proto = "ListenStream"
			case "udp":
				proto = "ListenDatagram"
			default:
				return fmt.Errorf("unrecognized protocol: %v", sap.Protocol)
			}
			sockopts = append(sockopts, unit.NewUnitOption("Socket", proto, fmt.Sprintf("%v", sap.Port)))
		}

		file, err := os.OpenFile(SocketUnitPath(p.Root, appName), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return errwrap.Wrap(errors.New("failed to create socket file"), err)
		}
		defer file.Close()

		if _, err = io.Copy(file, unit.Serialize(sockopts)); err != nil {
			return errwrap.Wrap(errors.New("failed to write socket unit file"), err)
		}

		if err = os.Symlink(path.Join("..", SocketUnitName(appName)), SocketWantPath(p.Root, appName)); err != nil {
			return errwrap.Wrap(errors.New("failed to link socket want"), err)
		}

		opts = append(opts, unit.NewUnitOption("Unit", "Requires", SocketUnitName(appName)))
	}

	opts = append(opts, unit.NewUnitOption("Unit", "Requires", InstantiatedPrepareAppUnitName(appName)))
	opts = append(opts, unit.NewUnitOption("Unit", "After", InstantiatedPrepareAppUnitName(appName)))

	file, err := os.OpenFile(ServiceUnitPath(p.Root, appName), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errwrap.Wrap(errors.New("failed to create service unit file"), err)
	}
	defer file.Close()

	if _, err = io.Copy(file, unit.Serialize(opts)); err != nil {
		return errwrap.Wrap(errors.New("failed to write service unit file"), err)
	}

	if err = os.Symlink(path.Join("..", ServiceUnitName(appName)), ServiceWantPath(p.Root, appName)); err != nil {
		return errwrap.Wrap(errors.New("failed to link service want"), err)
	}

	if flavor == "kvm" {
		// bind mount all shared volumes from /mnt/volumeName (we don't use mechanism for bind-mounting given by nspawn)
		err := AppToSystemdMountUnits(common.Stage1RootfsPath(p.Root), appName, p.Manifest.Volumes, ra, UnitsDir)
		if err != nil {
			return errwrap.Wrap(errors.New("failed to prepare mount units"), err)
		}

	}

	if err = writeAppReaper(p, appName.String()); err != nil {
		return errwrap.Wrap(fmt.Errorf("failed to write app %q reaper service", appName), err)
	}

	return nil
}

// writeEnvFile creates an environment file for given app name, the minimum
// required environment variables by the appc spec will be set to sensible
// defaults here if they're not provided by env.
func writeEnvFile(p *stage1commontypes.Pod, env types.Environment, appName types.ACName, privateUsers string) error {
	ef := bytes.Buffer{}

	for dk, dv := range defaultEnv {
		if _, exists := env.Get(dk); !exists {
			fmt.Fprintf(&ef, "%s=%s\000", dk, dv)
		}
	}

	for _, e := range env {
		fmt.Fprintf(&ef, "%s=%s\000", e.Name, e.Value)
	}

	uidRange := uid.NewBlankUidRange()
	if err := uidRange.Deserialize([]byte(privateUsers)); err != nil {
		return err
	}

	envFilePath := EnvFilePath(p.Root, appName)
	if err := ioutil.WriteFile(envFilePath, ef.Bytes(), 0644); err != nil {
		return err
	}

	if uidRange.Shift != 0 && uidRange.Count != 0 {
		if err := os.Chown(envFilePath, int(uidRange.Shift), int(uidRange.Shift)); err != nil {
			return err
		}
	}

	return nil
}

// PodToSystemd creates the appropriate systemd service unit files for
// all the constituent apps of the Pod
func PodToSystemd(p *stage1commontypes.Pod, interactive bool, flavor string, privateUsers string) error {

	for i := range p.Manifest.Apps {
		ra := &p.Manifest.Apps[i]
		if err := appToSystemd(p, ra, interactive, flavor, privateUsers); err != nil {
			return errwrap.Wrap(fmt.Errorf("failed to transform app %q into systemd service", ra.Name), err)
		}
	}
	return nil
}

// appToNspawnArgs transforms the given app manifest, with the given associated
// app name, into a subset of applicable systemd-nspawn argument
func appToNspawnArgs(p *stage1commontypes.Pod, ra *schema.RuntimeApp) ([]string, error) {
	var args []string
	appName := ra.Name
	app := ra.App

	sharedVolPath := common.SharedVolumesPath(p.Root)
	if err := os.MkdirAll(sharedVolPath, sharedVolPerm); err != nil {
		return nil, errwrap.Wrap(errors.New("could not create shared volumes directory"), err)
	}
	if err := os.Chmod(sharedVolPath, sharedVolPerm); err != nil {
		return nil, errwrap.Wrap(fmt.Errorf("could not change permissions of %q", sharedVolPath), err)
	}

	vols := make(map[types.ACName]types.Volume)
	for _, v := range p.Manifest.Volumes {
		vols[v.Name] = v
	}

	mounts := GenerateMounts(ra, vols)
	for _, m := range mounts {
		vol := vols[m.Volume]

		if vol.Kind == "empty" {
			p := filepath.Join(sharedVolPath, vol.Name.String())
			if err := os.MkdirAll(p, sharedVolPerm); err != nil {
				return nil, errwrap.Wrap(fmt.Errorf("could not create shared volume %q", vol.Name), err)
			}
			if err := os.Chown(p, *vol.UID, *vol.GID); err != nil {
				return nil, errwrap.Wrap(fmt.Errorf("could not change owner of %q", p), err)
			}
			mod, err := strconv.ParseUint(*vol.Mode, 8, 32)
			if err != nil {
				return nil, errwrap.Wrap(fmt.Errorf("invalid mode %q for volume %q", *vol.Mode, vol.Name), err)
			}
			if err := os.Chmod(p, os.FileMode(mod)); err != nil {
				return nil, errwrap.Wrap(fmt.Errorf("could not change permissions of %q", p), err)
			}
		}

		opt := make([]string, 4)

		if IsMountReadOnly(vol, app.MountPoints) {
			opt[0] = "--bind-ro="
		} else {
			opt[0] = "--bind="
		}

		switch vol.Kind {
		case "host":
			opt[1] = vol.Source
		case "empty":
			absRoot, err := filepath.Abs(p.Root)
			if err != nil {
				return nil, errwrap.Wrap(errors.New("cannot get pod's root absolute path"), err)
			}
			opt[1] = filepath.Join(common.SharedVolumesPath(absRoot), vol.Name.String())
		default:
			return nil, fmt.Errorf(`invalid volume kind %q. Must be one of "host" or "empty"`, vol.Kind)
		}
		opt[2] = ":"
		opt[3] = filepath.Join(common.RelAppRootfsPath(appName), m.Path)

		args = append(args, strings.Join(opt, ""))
	}

	for _, i := range app.Isolators {
		switch v := i.Value().(type) {
		case types.LinuxCapabilitiesSet:
			var caps []string
			// TODO: cleanup the API on LinuxCapabilitiesSet to give strings easily.
			for _, c := range v.Set() {
				caps = append(caps, string(c))
			}
			if i.Name == types.LinuxCapabilitiesRetainSetName {
				capList := strings.Join(caps, ",")
				args = append(args, "--capability="+capList)
			}
		}
	}

	return args, nil
}

// PodToNspawnArgs renders a prepared Pod as a systemd-nspawn
// argument list ready to be executed
func PodToNspawnArgs(p *stage1commontypes.Pod) ([]string, error) {
	args := []string{
		"--uuid=" + p.UUID.String(),
		"--machine=" + GetMachineID(p),
		"--directory=" + common.Stage1RootfsPath(p.Root),
	}

	for i := range p.Manifest.Apps {
		aa, err := appToNspawnArgs(p, &p.Manifest.Apps[i])
		if err != nil {
			return nil, err
		}
		args = append(args, aa...)
	}

	return args, nil
}

// GetFlavor populates a flavor string based on the flavor itself and respectively the systemd version
func GetFlavor(p *stage1commontypes.Pod) (flavor string, systemdVersion string, err error) {
	flavor, err = os.Readlink(filepath.Join(common.Stage1RootfsPath(p.Root), "flavor"))
	if err != nil {
		return "", "", errwrap.Wrap(errors.New("unable to determine stage1 flavor"), err)
	}

	if flavor == "host" {
		// This flavor does not contain systemd, so don't return systemdVersion
		return flavor, "", nil
	}

	systemdVersionBytes, err := ioutil.ReadFile(filepath.Join(common.Stage1RootfsPath(p.Root), "systemd-version"))
	if err != nil {
		return "", "", errwrap.Wrap(errors.New("unable to determine stage1's systemd version"), err)
	}
	systemdVersion = strings.Trim(string(systemdVersionBytes), " \n")
	return flavor, systemdVersion, nil
}

// GetAppHashes returns a list of hashes of the apps in this pod
func GetAppHashes(p *stage1commontypes.Pod) []types.Hash {
	var names []types.Hash
	for _, a := range p.Manifest.Apps {
		names = append(names, a.Image.ID)
	}

	return names
}

// GetMachineID returns the machine id string of the pod to be passed to
// systemd-nspawn
func GetMachineID(p *stage1commontypes.Pod) string {
	return "rkt-" + p.UUID.String()
}
