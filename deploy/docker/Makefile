# Copyright 2016 The Kubernetes Authors All rights reserved.
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

REGISTRY?=gcr.io/google_containers
ARCH?=amd64
TEMP_DIR:=$(shell mktemp -d)

all: build
build:

ifndef VERSION
    $(error VERSION is undefined)
endif

	cp -r ./* $(TEMP_DIR)
	cp -r ../addons ./static-pods $(TEMP_DIR)
	cp ../../out/localkube $(TEMP_DIR)
	chmod +x $(TEMP_DIR)/localkube

	docker build -t $(REGISTRY)/localkube-$(ARCH):$(VERSION) $(TEMP_DIR)
