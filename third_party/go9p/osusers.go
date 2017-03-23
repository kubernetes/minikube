// Copyright 2009 The Go9p Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go9p

import "sync"

var once sync.Once

type osUser struct {
	uid int
}

type osUsers struct {
	users  map[int]*osUser
	groups map[int]*osGroup
	sync.Mutex
}

// Simple Users implementation that fakes looking up users and groups
// by uid only. The names and groups memberships are empty
var OsUsers *osUsers

func (u *osUser) Name() string { return "" }

func (u *osUser) Id() int { return u.uid }

func (u *osUser) Groups() []Group { return nil }

func (u *osUser) IsMember(g Group) bool { return false }

type osGroup struct {
	gid int
}

func (g *osGroup) Name() string { return "" }

func (g *osGroup) Id() int { return g.gid }

func (g *osGroup) Members() []User { return nil }

func initOsusers() {
	OsUsers = new(osUsers)
	OsUsers.users = make(map[int]*osUser)
	OsUsers.groups = make(map[int]*osGroup)
}

func (up *osUsers) Uid2User(uid int) User {
	once.Do(initOsusers)
	OsUsers.Lock()
	defer OsUsers.Unlock()
	user, present := OsUsers.users[uid]
	if present {
		return user
	}

	user = new(osUser)
	user.uid = uid
	OsUsers.users[uid] = user
	return user
}

func (up *osUsers) Uname2User(uname string) User {
	// unimplemented
	return nil
}

func (up *osUsers) Gid2Group(gid int) Group {
	once.Do(initOsusers)
	OsUsers.Lock()
	group, present := OsUsers.groups[gid]
	if present {
		OsUsers.Unlock()
		return group
	}

	group = new(osGroup)
	group.gid = gid
	OsUsers.groups[gid] = group
	OsUsers.Unlock()
	return group
}

func (up *osUsers) Gname2Group(gname string) Group {
	return nil
}
