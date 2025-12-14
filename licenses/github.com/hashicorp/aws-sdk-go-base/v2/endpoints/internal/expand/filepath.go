// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package expand

import (
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/go-homedir"
)

func FilePaths(in []string) ([]string, error) {
	var errs *multierror.Error
	result := make([]string, 0, len(in))
	for _, v := range in {
		p, err := FilePath(v)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		result = append(result, p)
	}
	return result, errs.ErrorOrNil()
}

func FilePath(in string) (s string, err error) {
	e := os.ExpandEnv(in)
	s, err = homedir.Expand(e)
	return
}
