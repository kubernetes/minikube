// Copyright 2014 Google Inc. All Rights Reserved.
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

package riemann

import (
	"net/url"
	"runtime"
	"strconv"
	"sync"

	"time"

	riemann_api "github.com/bigdatadev/goryman"
	"github.com/golang/glog"
	"k8s.io/heapster/metrics/core"
)

// Abstracted for testing: this package works against any client that obeys the
// interface contract exposed by the goryman Riemann client

type riemannClient interface {
	Connect() error
	Close() error
	SendEvent(e *riemann_api.Event) error
}

type riemannSink struct {
	client riemannClient
	config riemannConfig
	sync.RWMutex
}

type riemannConfig struct {
	host  string
	ttl   float32
	state string
	tags  []string
}

const (
	// Maximum number of riemann Events to be sent in one batch.
	maxSendBatchSize = 10000
	max_retries      = 2
)

func CreateRiemannSink(uri *url.URL) (core.DataSink, error) {
	c := riemannConfig{
		host:  "riemann-heapster:5555",
		ttl:   60.0,
		state: "",
		tags:  make([]string, 0),
	}
	if len(uri.Host) > 0 {
		c.host = uri.Host
	}
	options := uri.Query()
	if len(options["ttl"]) > 0 {
		var ttl, err = strconv.ParseFloat(options["ttl"][0], 32)
		if err != nil {
			return nil, err
		}
		c.ttl = float32(ttl)
	}
	if len(options["state"]) > 0 {
		c.state = options["state"][0]
	}
	if len(options["tags"]) > 0 {
		c.tags = options["tags"]
	}

	glog.Infof("Riemann sink URI: '%+v', host: '%+v', options: '%+v', ", uri, c.host, options)
	rs := &riemannSink{
		client: nil,
		config: c,
	}

	err := rs.setupRiemannClient()
	if err != nil {
		glog.Warningf("Riemann sink not connected: %v", err)
		// Warn but return the sink.
	}
	return rs, nil
}

func (rs *riemannSink) setupRiemannClient() error {
	client := riemann_api.NewGorymanClient(rs.config.host)
	runtime.SetFinalizer(client, func(c riemannClient) { c.Close() })
	err := client.Connect()
	if err != nil {
		return err
	}
	rs.client = client
	return nil
}

// Return a user-friendly string describing the sink
func (sink *riemannSink) Name() string {
	return "Riemann Sink"
}

func (sink *riemannSink) Stop() {
	// nothing needs to be done.
}

// ExportData Send a collection of Timeseries to Riemann
func (sink *riemannSink) ExportData(dataBatch *core.DataBatch) {
	sink.Lock()
	defer sink.Unlock()

	if sink.client == nil {
		if err := sink.setupRiemannClient(); err != nil {
			glog.Warningf("Riemann sink not connected: %v", err)
			return
		}
	}

	dataEvents := make([]riemann_api.Event, 0, 0)
	for _, metricSet := range dataBatch.MetricSets {
		for metricName, metricValue := range metricSet.MetricValues {
			if value := metricValue.GetValue(); value != nil {
				event := riemann_api.Event{
					Time:        dataBatch.Timestamp.Unix(),
					Service:     metricName,
					Host:        metricSet.Labels[core.LabelHostname.Key],
					Description: "", //no description - waste of bandwidth.
					Attributes:  metricSet.Labels,
					Metric:      value,
					Ttl:         sink.config.ttl,
					State:       sink.config.state,
					Tags:        sink.config.tags,
				}

				dataEvents = append(dataEvents, event)
				if len(dataEvents) >= maxSendBatchSize {
					sink.sendData(dataEvents)
					dataEvents = make([]riemann_api.Event, 0, 0)
				}
			}
		}
		for _, metric := range metricSet.LabeledMetrics {
			if value := metric.GetValue(); value != nil {
				labels := make(map[string]string)
				for k, v := range metricSet.Labels {
					labels[k] = v
				}
				for k, v := range metric.Labels {
					labels[k] = v
				}
				event := riemann_api.Event{
					Time:        dataBatch.Timestamp.Unix(),
					Service:     metric.Name,
					Host:        metricSet.Labels[core.LabelHostname.Key],
					Description: "", //no description - waste of bandwidth.
					Attributes:  labels,
					Metric:      value,
					Ttl:         sink.config.ttl,
					State:       sink.config.state,
					Tags:        sink.config.tags,
				}
				dataEvents = append(dataEvents, event)
				if len(dataEvents) >= maxSendBatchSize {
					sink.sendData(dataEvents)
					dataEvents = make([]riemann_api.Event, 0, 0)
				}
			}
		}
	}

	if len(dataEvents) >= 0 {
		sink.sendData(dataEvents)
	}
}

func (sink *riemannSink) sendData(dataEvents []riemann_api.Event) {
	if sink.client == nil {
		return
	}

	start := time.Now()
	for _, event := range dataEvents {
		var err error = nil
		for try := 0; try < max_retries; try++ {
			err = sink.client.SendEvent(&event)
			if err == nil {
				break
			}
		}
		if err != nil {
			glog.V(2).Infof("Failed sending event to Riemann: %+v: %+v", event, err)
			// Let's reconnect with the next export.
			// Assumes that this happens under a lock.
			sink.client.Close()
			sink.client = nil
		}
	}
	end := time.Now()
	glog.V(4).Info("Exported %d data to riemann in %s", len(dataEvents), end.Sub(start))
}
