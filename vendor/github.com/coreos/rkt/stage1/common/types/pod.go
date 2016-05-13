// Copyright 2015 The rkt Authors
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

package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/coreos/rkt/common"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/hashicorp/errwrap"
)

// Pod encapsulates a PodManifest and ImageManifests
type Pod struct {
	Root               string // root directory where the pod will be located
	UUID               types.UUID
	Manifest           *schema.PodManifest
	Images             map[string]*schema.ImageManifest
	MetadataServiceURL string
	Networks           []string
}

// LoadPod loads a Pod Manifest (as prepared by stage0) and
// its associated Application Manifests, under $root/stage1/opt/stage1/$apphash
func LoadPod(root string, uuid *types.UUID) (*Pod, error) {
	p := &Pod{
		Root:   root,
		UUID:   *uuid,
		Images: make(map[string]*schema.ImageManifest),
	}

	buf, err := ioutil.ReadFile(common.PodManifestPath(p.Root))
	if err != nil {
		return nil, errwrap.Wrap(errors.New("failed reading pod manifest"), err)
	}

	pm := &schema.PodManifest{}
	if err := json.Unmarshal(buf, pm); err != nil {
		return nil, errwrap.Wrap(errors.New("failed unmarshalling pod manifest"), err)
	}
	p.Manifest = pm

	for i, app := range p.Manifest.Apps {
		ampath := common.ImageManifestPath(p.Root, app.Name)
		buf, err := ioutil.ReadFile(ampath)
		if err != nil {
			return nil, errwrap.Wrap(fmt.Errorf("failed reading app manifest %q", ampath), err)
		}

		am := &schema.ImageManifest{}
		if err = json.Unmarshal(buf, am); err != nil {
			return nil, errwrap.Wrap(fmt.Errorf("failed unmarshalling app manifest %q", ampath), err)
		}

		if _, ok := p.Images[app.Name.String()]; ok {
			return nil, fmt.Errorf("got multiple definitions for app: %v", app.Name)
		}
		if app.App == nil {
			p.Manifest.Apps[i].App = am.App
		}
		p.Images[app.Name.String()] = am
	}

	return p, nil
}
