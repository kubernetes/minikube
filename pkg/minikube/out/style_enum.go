/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package out

// StyleEnum is an enumeration of Style
type StyleEnum int

// All the Style constants available
const (
	AddonDisable = iota
	AddonEnable
	Caching
	Celebrate
	Celebration
	Check
	Command
	Conflict
	Confused
	Connectivity
	Containerd
	ContainerRuntime
	Copying
	CRIO
	Deleted
	DeletingHost
	Docker
	Documentation
	DryRun
	Embarrassed
	Empty
	Enabling
	FailureType
	FatalType
	FileDownload
	Fileserver
	Happy StyleEnum
	HealthCheck
	Internet
	ISODownload
	Issue
	Issues
	Kubectl
	Launch
	LogEntry
	Meh
	Mounting
	MountOptions
	New
	Notice
	Option
	Pause
	Permissions
	Provisioner
	Provisioning
	Pulling
	Ready
	Resetting
	Restarting
	Running
	Sad
	Shrug
	Shutdown
	Sparkle
	StartingNone
	StartingVM
	Stopped
	Stopping
	SuccessType
	ThumbsDown
	ThumbsUp
	Tip
	Unmount
	Unpause
	URL
	Usage
	Verifying
	VerifyingNoLine
	Waiting
	WaitingPods
	Warning
	Workaround
)
