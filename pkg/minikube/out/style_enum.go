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
	Happy StyleEnum = iota
	SuccessType
	FailureType
	Celebration
	Conflict
	FatalType
	Notice
	Ready
	Running
	Provisioning
	Restarting
	Stopping
	Stopped
	Warning
	Waiting
	WaitingPods
	Usage
	Launch
	Sad
	ThumbsUp
	ThumbsDown
	Option
	Command
	LogEntry
	Deleted
	URL
	Documentation
	Issues
	Issue
	Check
	ISODownload
	FileDownload
	Caching
	StartingVM
	StartingNone
	Provisioner
	Resetting
	DeletingHost
	Copying
	Connectivity
	Confused
	Internet
	Mounting
	Celebrate
	ContainerRuntime
	Docker
	CRIO
	Containerd
	Permissions
	Enabling
	Shutdown
	Pulling
	Verifying
	VerifyingNoLine
	Kubectl
	Meh
	Embarrassed
	Tip
	Unmount
	MountOptions
	Fileserver
	Empty
	Workaround
	Sparkle
	Pause
	Unpause
	DryRun
	AddonEnable
	AddonDisable
	Shrug
)
