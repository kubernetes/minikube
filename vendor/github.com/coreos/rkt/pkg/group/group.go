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

package group

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
)

const (
	groupFilePath = "/etc/group"
)

// Group represents an entry in the group file.
type Group struct {
	Name  string
	Pass  string
	Gid   int
	Users []string
}

// LookupGid reads the group file and returns the gid of the group
// specified by groupName.
func LookupGid(groupName string) (gid int, err error) {
	groups, err := parseGroupFile(groupFilePath)
	if err != nil {
		return -1, errwrap.Wrap(fmt.Errorf("error parsing %q file", groupFilePath), err)
	}

	group, ok := groups[groupName]
	if !ok {
		return -1, fmt.Errorf("%q group not found", groupName)
	}

	return group.Gid, nil
}

func parseGroupFile(path string) (group map[string]Group, err error) {
	groupFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer groupFile.Close()

	return parseGroups(groupFile)
}

func parseGroups(r io.Reader) (group map[string]Group, err error) {
	s := bufio.NewScanner(r)
	out := make(map[string]Group)

	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}

		text := s.Text()
		if text == "" {
			continue
		}

		p := Group{}
		parseGroupLine(text, &p)

		out[p.Name] = p
	}

	return out, nil
}

func parseGroupLine(line string, group *Group) {
	const (
		NameIdx = iota
		PassIdx
		GidIdx
		UsersIdx
	)

	if line == "" {
		return
	}

	splits := strings.Split(line, ":")
	if len(splits) < 4 {
		return
	}

	group.Name = splits[NameIdx]
	group.Pass = splits[PassIdx]
	group.Gid, _ = strconv.Atoi(splits[GidIdx])

	u := splits[UsersIdx]
	if u != "" {
		group.Users = strings.Split(u, ",")
	} else {
		group.Users = []string{}
	}
}
