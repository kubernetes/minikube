/*
   Copyright 2015-2016 Red Hat, Inc. and/or its affiliates
   and other contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package metrics

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func (c *Client) sendRoutine() {
	for {
		select {
		case pr, open := <-c.pool:
			if !open {
				return
			}
			resp, err := c.client.Do(pr.req)
			pr.rChan <- &poolResponse{err, resp}
		}
	}
}

// ConvertToFloat64 Return float64 from most numeric types
func ConvertToFloat64(v interface{}) (float64, error) {
	switch i := v.(type) {
	case float64:
		return float64(i), nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int16:
		return float64(i), nil
	case int8:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint16:
		return float64(i), nil
	case uint8:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		f, err := strconv.ParseFloat(i, 64)
		if err != nil {
			return math.NaN(), err
		}
		return f, err
	default:
		return math.NaN(), fmt.Errorf("Cannot convert %s to float64", i)
	}
}

// UnixMilli Returns milliseconds since epoch
func UnixMilli(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

// Prepend Helper function to insert modifier in the beginning of slice
func prepend(slice []Modifier, a ...Modifier) []Modifier {
	p := make([]Modifier, 0, len(slice)+len(a))
	p = append(p, a...)
	p = append(p, slice...)
	return p
}
