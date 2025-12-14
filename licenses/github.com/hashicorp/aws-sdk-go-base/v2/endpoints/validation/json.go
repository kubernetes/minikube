// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// DuplicateKeyError is returned when duplicate key names are detected
// inside a JSON object
type DuplicateKeyError struct {
	path []string
	key  string
}

func (e *DuplicateKeyError) Error() string {
	return fmt.Sprintf(`duplicate key "%s"`, strings.Join(append(e.path, e.key), "."))
}

// JSONNoDuplicateKeys verifies the provided JSON object contains
// no duplicated keys
//
// The function expects a single JSON object, and will error prior to
// checking for duplicate keys should an invalid input be provided.
func JSONNoDuplicateKeys(s string) error {
	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return fmt.Errorf("unmarshaling input: %w", err)
	}

	dec := json.NewDecoder(strings.NewReader(s))
	return checkToken(dec, nil)
}

// checkToken walks a JSON object checking for duplicated keys
//
// The function is called recursively on the value of each key
// inside and object, or item inside an array.
//
// Adapted from: https://stackoverflow.com/a/50109335
func checkToken(dec *json.Decoder, path []string) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	delim, ok := t.(json.Delim)
	if !ok {
		// non-delimiter, nothing to do
		return nil
	}

	var dupErrs []error
	switch delim {
	case '{':
		keys := make(map[string]bool)
		for dec.More() {
			// Get the field key
			t, err := dec.Token()
			if err != nil {
				return err
			}
			key := t.(string)

			if keys[key] {
				// Duplicate found
				dupErrs = append(dupErrs, &DuplicateKeyError{path: path, key: key})
			}
			keys[key] = true

			// Check the keys value
			if err := checkToken(dec, append(path, key)); err != nil {
				dupErrs = append(dupErrs, err)
			}
		}

		// consume trailing "}"
		_, err := dec.Token()
		if err != nil {
			return err
		}
	case '[':
		i := 0
		for dec.More() {
			// Check each items value
			if err := checkToken(dec, append(path, strconv.Itoa(i))); err != nil {
				dupErrs = append(dupErrs, err)
			}
			i++
		}

		// consume trailing "]"
		_, err := dec.Token()
		if err != nil {
			return err
		}
	}

	return errors.Join(dupErrs...)
}
