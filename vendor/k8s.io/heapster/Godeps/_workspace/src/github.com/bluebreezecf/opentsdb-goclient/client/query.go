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
// Package client defines the client and the corresponding
// rest api implementaion of OpenTSDB.
//
// query.go contains the structs and methods for the implementation of /api/query.
//
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// QueryParam is the structure used to hold
// the querying parameters when calling /api/query.
// Each attributes in QueryParam matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/query/index.html).
//
type QueryParam struct {
	// The start time for the query. This can be a relative or absolute timestamp.
	// The data type can only be string, int, or int64.
	// The value is required with non-zero value of the target type.
	Start interface{} `json:"start"`

	// An end time for the query. If not supplied, the TSD will assume the local
	// system time on the server. This may be a relative or absolute timestamp.
	// The data type can only be string, or int64.
	// The value is optional.
	End interface{} `json:"end,omitempty"`

	// One or more sub queries used to select the time series to return.
	// These may be metric m or TSUID tsuids queries
	// The value is required with at least one element
	Queries []SubQuery `json:"queries"`

	// An optional value is used to show whether or not to return annotations with a query.
	// The default is to return annotations for the requested timespan but this flag can disable the return.
	// This affects both local and global notes and overrides globalAnnotations
	NoAnnotations bool `json:"noAnnotations,omitempty"`

	// An optional value is used to show whether or not the query should retrieve global
	// annotations for the requested timespan.
	GlobalAnnotations bool `json:"globalAnnotations,omitempty"`

	// An optional value is used to show whether or not to output data point timestamps in milliseconds or seconds.
	// If this flag is not provided and there are multiple data points within a second,
	// those data points will be down sampled using the query's aggregation function.
	MsResolution bool `json:"msResolution,omitempty"`

	// An optional value is used to show whether or not to output the TSUIDs associated with timeseries in the results.
	// If multiple time series were aggregated into one set, multiple TSUIDs will be returned in a sorted manner.
	ShowTSUIDs bool `json:"showTSUIDs,omitempty"`
}

func (query *QueryParam) String() string {
	content, _ := json.Marshal(query)
	return string(content)
}

// SubQuery is the structure used to hold
// the subquery parameters when calling /api/query.
// Each attributes in SubQuery matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/query/index.html).
//
type SubQuery struct {
	// The name of an aggregation function to use.
	// The value is required with non-empty one in the range of
	// the response of calling /api/aggregators.
	Aggregator string `json:"aggregator"`

	// The name of a metric stored in the system.
	// The value is reqiured with non-empty value.
	Metric string `json:"metric"`

	// An optional value is used to show whether or not the data should be
	// converted into deltas before returning. This is useful if the metric is a
	// continously incrementing counter and you want to view the rate of change between data points.
	Rate bool `json:"rate,omitempty"`

	// rateOptions represents monotonically increasing counter handling options.
	// The value is optional.
	// Currently there is only three kind of value can be set to this map:
	// Only three keys can be set into the rateOption parameter of the QueryParam is
	// QueryRateOptionCounter (value type is bool),  QueryRateOptionCounterMax (value type is int,int64)
	// QueryRateOptionResetValue (value type is int,int64)
	RateParams map[string]interface{} `json:"rateOptions,omitempty"`

	// An optional value downsampling function to reduce the amount of data returned.
	Downsample string `json:"downsample,omitempty"`

	// An optional value to drill down to specific timeseries or group results by tag,
	// supply one or more map values in the same format as the query string. Tags are converted to filters in 2.2.
	// Note that if no tags are specified, all metrics in the system will be aggregated into the results.
	// It will be deprecated in OpenTSDB 2.2.
	Tags map[string]string `json:"tags,omitempty"`

	// An optional value used to filter the time series emitted in the results.
	// Note that if no filters are specified, all time series for the given
	// metric will be aggregated into the results.
	Fiters []Filter `json:"filters,omitempty"`
}

// Filter is the structure used to hold the filter parameters when calling /api/query.
// Each attributes in Filter matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/query/index.html).
//
type Filter struct {
	// The name of the filter to invoke. The value is required with a non-empty
	// value in the range of calling /api/config/filters.
	Type string `json:"type"`

	// The tag key to invoke the filter on, required with a non-empty value
	Tagk string `json:"tagk"`

	// The filter expression to evaluate and depends on the filter being used, required with a non-empty value
	FilterExp string `json:"filter"`

	// An optional value to show whether or not to group the results by each value matched by the filter.
	// By default all values matching the filter will be aggregated into a single series.
	GroupBy bool `json:"groupBy"`
}

// QueryResponse acts as the implementation of Response in the /api/query scene.
// It holds the status code and the response values defined in the
// (http://opentsdb.net/docs/build/html/api_http/query/index.html).
//
type QueryResponse struct {
	StatusCode    int
	QueryRespCnts []QueryRespItem        `json:"queryRespCnts"`
	ErrorMsg      map[string]interface{} `json:"error"`
}

func (queryResp *QueryResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(queryResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (queryResp *QueryResponse) SetStatus(code int) {
	queryResp.StatusCode = code
}

func (queryResp *QueryResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		originRespStr := string(respCnt)
		var respStr string
		if queryResp.StatusCode == 200 && strings.Contains(originRespStr, "[") && strings.Contains(originRespStr, "]") {
			respStr = fmt.Sprintf("{%s:%s}", `"queryRespCnts"`, originRespStr)
		} else {
			respStr = originRespStr
		}
		return json.Unmarshal([]byte(respStr), &queryResp)
	}
}

// QueryRespItem acts as the implementation of Response in the /api/query scene.
// It holds the response item defined in the
// (http://opentsdb.net/docs/build/html/api_http/query/index.html).
//
type QueryRespItem struct {
	// Name of the metric retreived for the time series
	Metric string `json:"metric"`

	// A list of tags only returned when the results are for a single time series.
	// If results are aggregated, this value may be null or an empty map
	Tags map[string]string `json:"tags"`

	// If more than one timeseries were included in the result set, i.e. they were aggregated,
	// this will display a list of tag names that were found in common across all time series.
	// Note that: Api Doc uses 'aggreatedTags', but actual response uses 'aggregateTags'
	AggregatedTags []string `json:"aggregateTags"`

	// Retrieved data points after being processed by the aggregators. Each data point consists
	// of a timestamp and a value, the format determined by the serializer.
	// For the JSON serializer, the timestamp will always be a Unix epoch style integer followed
	// by the value as an integer or a floating point.
	// For example, the default output is "dps"{"<timestamp>":<value>}.
	// By default the timestamps will be in seconds. If the msResolution flag is set, then the
	// timestamps will be in milliseconds.
	Dps map[string]interface{} `json:"dps"`

	// If the query retrieved annotations for timeseries over the requested timespan, they will
	// be returned in this group. Annotations for every timeseries will be merged into one set
	// and sorted by start_time. Aggregator functions do not affect annotations, all annotations
	// will be returned for the span.
	// The value is optional.
	Annotations []Annotation `json:"annotations,omitempty"`

	// If requested by the user, the query will scan for global annotations during
	// the timespan and the results returned in this group.
	// The value is optional.
	GlobalAnnotations []Annotation `json:"globalAnnotations,omitempty"`
}

func (c *clientImpl) Query(param QueryParam) (*QueryResponse, error) {
	if !isValidQueryParam(&param) {
		return nil, errors.New("The given query param is invalid.\n")
	}
	queryEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, QueryPath)
	reqBodyCnt, err := getQueryBodyContents(&param)
	if err != nil {
		return nil, err
	}
	queryResp := QueryResponse{}
	if err = c.sendRequest(PostMethod, queryEndpoint, reqBodyCnt, &queryResp); err != nil {
		return nil, err
	}
	return &queryResp, nil
}

func getQueryBodyContents(param *QueryParam) (string, error) {
	result, err := json.Marshal(param)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to marshal query param: %v\n", err))
	}
	return string(result), nil
}

func isValidQueryParam(param *QueryParam) bool {
	if param.Queries == nil || len(param.Queries) == 0 {
		return false
	}
	if !isValidTimePoint(param.Start) {
		return false
	}
	for _, query := range param.Queries {
		if len(query.Aggregator) == 0 || len(query.Metric) == 0 {
			return false
		}
		for k, _ := range query.RateParams {
			if k != QueryRateOptionCounter && k != QueryRateOptionCounterMax && k != QueryRateOptionResetValue {
				return false
			}
		}
	}
	return true
}

func isValidTimePoint(timePoint interface{}) bool {
	if timePoint == nil {
		return false
	}
	switch v := timePoint.(type) {
	case int:
		if v <= 0 {
			return false
		}
	case int64:
		if v <= 0 {
			return false
		}
	case string:
		if v == "" {
			return false
		}

	default:
		return false
	}
	return true
}
