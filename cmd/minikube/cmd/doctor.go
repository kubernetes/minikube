/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/minikube/cmd/minikube/cmd/flags"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/doctor"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/style"
)

// driverStatusTimeout matches the timeout `minikube start` uses when it
// probes drivers.
const driverStatusTimeout = 20 * time.Second

var doctorOutput string

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose problems with a minikube profile and its host",
	Long: `Runs read-only checks on the profile, the driver, the cluster and its
resources, and suggests a fix for each problem found. It works even when the
profile does not exist or the cluster is not running.`,
	Run: func(_ *cobra.Command, _ []string) {
		options := flags.CommandOptions()
		profile := ClusterFlagValue()

		h := doctor.Host{
			LoadConfig: func(name string) (*config.ClusterConfig, error) { return config.Load(name) },
			ListProfiles: func() ([]*config.Profile, []*config.Profile, error) {
				return config.ListProfiles()
			},
			DriverStatus: func(name string) (registry.State, bool) {
				def := registry.Driver(name)
				if def.Empty() || def.Status == nil {
					return registry.State{}, false
				}
				ch := make(chan registry.State, 1)
				go func() { ch <- def.Status(options) }()
				select {
				case st := <-ch:
					return st, true
				case <-time.After(driverStatusTimeout):
					return registry.State{
						Installed: true,
						Error:     fmt.Errorf("%s did not respond within %s", name, driverStatusTimeout),
						Fix:       "Restart " + name + " and try again",
					}, true
				}
			},
			ClusterStatus: func(cc *config.ClusterConfig) ([]*cluster.Status, error) {
				api, err := machine.NewAPIClient(options)
				if err != nil {
					return nil, err
				}
				defer api.Close()
				return cluster.GetStatus(api, cc)
			},
			LookPath:       exec.LookPath,
			CurrentContext: func() (string, error) { return kubeconfig.GetCurrentContext() },
		}

		report := doctor.Run(profile, h)

		switch doctorOutput {
		case "json":
			data, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				exit.Error(reason.InternalJSONMarshal, "marshalling doctor report", err)
			}
			out.String(string(data) + "\n")
		case "text":
			printDoctorReport(report)
		default:
			exit.Message(reason.Usage, "invalid output format: {{.output}}. Valid values: 'text', 'json'", out.V{"output": doctorOutput})
		}

		if report.Count(doctor.Fail) > 0 {
			os.Exit(reason.ExProgramError)
		}
	},
}

func printDoctorReport(r doctor.Report) {
	out.Step(style.Check, "Profile: {{.profile}}", out.V{"profile": r.Profile})
	if r.Driver != "" {
		out.Infof("Driver: {{.driver}}, container runtime: {{.runtime}}, Kubernetes: {{.version}}",
			out.V{"driver": r.Driver, "runtime": r.Runtime, "version": r.Kubernetes})
	}

	var section doctor.Section
	for _, c := range r.Checks {
		if c.Section != section {
			section = c.Section
			out.Ln("")
			out.String(string(section) + "\n")
		}
		st := style.Success
		switch c.Status {
		case doctor.Warn:
			st = style.Warning
		case doctor.Fail:
			st = style.Failure
		case doctor.Skip:
			st = style.Empty
		}
		out.Step(st, "{{.name}}: {{.message}}", out.V{"name": c.Name, "message": c.Message})
		if c.Details != "" && c.Status != doctor.Pass {
			out.Infof("{{.details}}", out.V{"details": c.Details})
		}
		if c.Fix != "" {
			out.Styled(style.Tip, "{{.fix}}", out.V{"fix": c.Fix})
		}
	}

	out.Ln("")
	out.Step(style.Check, "{{.pass}} passed, {{.warn}} warnings, {{.fail}} failed, {{.skip}} skipped", out.V{
		"pass": r.Count(doctor.Pass), "warn": r.Count(doctor.Warn),
		"fail": r.Count(doctor.Fail), "skip": r.Count(doctor.Skip),
	})
}

func init() {
	doctorCmd.Flags().StringVarP(&doctorOutput, "output", "o", "text", "Output format: 'text' or 'json'")
}
