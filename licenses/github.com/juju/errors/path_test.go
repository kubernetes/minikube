// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package errors_test

import (
	"path/filepath"

	gc "gopkg.in/check.v1"

	"github.com/juju/errors"
)

type pathSuite struct{}

var _ = gc.Suite(&pathSuite{})

func (*pathSuite) TestTrimDefaultSet(c *gc.C) {
	c.Assert(errors.TrimDefault(), gc.Not(gc.Equals), "")
}

func (*pathSuite) TestTrimSourcePath(c *gc.C) {
	relativeImport := "github.com/foo/bar/rel.go"
	filename := filepath.Join(errors.TrimDefault(), relativeImport)
	c.Assert(errors.TrimSourcePath(filename), gc.Equals, relativeImport)

	absoluteImport := "/usr/share/foo/bar/abs.go"
	c.Assert(errors.TrimSourcePath(absoluteImport), gc.Equals, absoluteImport)
}

func (*pathSuite) TestSetSourceTrimPrefix(c *gc.C) {
	testPrefix := "/usr/share/"
	savePrefix := errors.SetSourceTrimPrefix(testPrefix)
	defer errors.SetSourceTrimPrefix(savePrefix)
	relative := "github.com/foo/bar/rel.go"
	filename := filepath.Join(testPrefix, relative)
	c.Assert(errors.TrimSourcePath(filename), gc.Equals, relative)
}
