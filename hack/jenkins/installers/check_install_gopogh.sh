#!/bin/bash

# Copyright 2023 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eux -o pipefail

# installing golang so we can go install gopogh
./installers/check_install_golang.sh "/usr/local" || true

# temporary: remove the old install of gopogh as it's taking priority over our current install, preventing updating
sudo rm -f /usr/local/bin/gopogh

go install github.com/medyagh/gopogh/cmd/gopogh@v0.29.0
