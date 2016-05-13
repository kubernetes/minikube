// Copyright 2015 opentsdb-goclient authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//
// Package config defines basic structure to hold the configure items used
// to initialize the opentsdb client.
//
package config

import (
	"net/http"
)

type OpenTSDBConfig struct {

	// The host of the target opentsdb, is a required non-empty string which is
	// in the format of ip:port without http:// prefix or a domain.
	OpentsdbHost string

	// A pointer of http.Tranport is used by the opentsdb client.
	// This value is optional, and if it is not set, client.DefaultTransport, which
	// enables tcp keepalive mode, will be used in the opentsdb client.
	Transport *http.Transport

	// The maximal number of datapoints which will be inserted into the opentsdb
	// via one calling of /api/put method.
	// This value is optional, and if it is not set, client.DefaultMaxPutPointsNum
	// will be used in the opentsdb client.
	MaxPutPointsNum int

	// The detect delta number of datapoints which will be used in client.Put()
	// to split a large group of datapoints into small batches.
	// This value is optional, and if it is not set, client.DefaultDetectDeltaNum
	// will be used in the opentsdb client.
	DetectDeltaNum int

	// The maximal body content length per /api/put method to insert datapoints
	// into opentsdb.
	// This value is optional, and if it is not set, client.DefaultMaxPutPointsNum
	// will be used in the opentsdb client.
	MaxContentLength int
}
