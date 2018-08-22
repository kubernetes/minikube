// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authn

import (
	"encoding/base64"
	"fmt"
)

// Basic implements Authenticator for basic authentication.
type Basic struct {
	Username string
	Password string
}

// Authorization implements Authenticator.
func (b *Basic) Authorization() (string, error) {
	delimited := fmt.Sprintf("%s:%s", b.Username, b.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(delimited))
	return fmt.Sprintf("Basic %s", encoded), nil
}
