/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package detect

import (
	"net/http"
	"os"
)

// IsMicrosoftWSL will return true if process is running in WSL in windows
// checking for WSL env var based on this https://github.com/microsoft/WSL/issues/423#issuecomment-608237689
// also based on https://github.com/microsoft/vscode/blob/90a39ba0d49d75e9a4d7e62a6121ad946ecebc58/resources/win32/bin/code.sh#L24
func IsMicrosoftWSL() bool {
	return os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSLPATH") != ""
}

// IsOnGCE determines whether minikube is currently running on GCE.
func IsOnGCE() bool {
	resp, err := http.Get("http://metadata.google.internal")
	if err != nil {
		return false
	}

	return resp.Header.Get("Metadata-Flavor") == "Google"
}
