## go-acl

[![Build status](https://ci.appveyor.com/api/projects/status/rbdyu7c39o2j0ru9?svg=true)](https://ci.appveyor.com/project/nathan-osman/go-acl)
[![GoDoc](https://godoc.org/github.com/hectane/go-acl?status.svg)](https://godoc.org/github.com/hectane/go-acl)
[![MIT License](http://img.shields.io/badge/license-MIT-9370d8.svg?style=flat)](http://opensource.org/licenses/MIT)

Manipulating ACLs (Access Control Lists) on Windows is difficult. go-acl wraps the Windows API functions that control access to objects, simplifying the process.

### Using the Package

To use the package add the following imports:

    import (
        "github.com/hectane/go-acl"
        "golang.org/x/sys/windows"
    )

### Examples

Probably the most commonly used function in this package is `Chmod`:

    if err := acl.Chmod("C:\\path\\to\\file.txt", 0755); err != nil {
        panic(err)
    }

To grant read access to user "Alice" and deny write access to user "Bob":

    if err := acl.Apply(
        "C:\\path\\to\\file.txt",
        false,
        false,
        acl.GrantName(windows.GENERIC_READ, "Alice"),
        acl.DenyName(windows.GENERIC_WRITE, "Bob"),
    ); err != nil {
        panic(err)
    }

### Using the API Directly

go-acl's `api` package exposes the individual Windows API functions that are used to manipulate ACLs. For example, to retrieve the current owner of a file:

    import (
        "github.com/hectane/go-acl/api"
        "golang.org/x/sys/windows"
    )

    var (
        owner   *windows.SID
        secDesc windows.Handle
    )
    err := api.GetNamedSecurityInfo(
        "C:\\path\\to\\file.txt",
        api.SE_FILE_OBJECT,
        api.OWNER_SECURITY_INFORMATION,
        &owner,
        nil,
        nil,
        nil,
        &secDesc,
    )
    if err != nil {
        panic(err)
    }
    defer windows.LocalFree(secDesc)

`owner` will then point to the SID for the owner of the file.
