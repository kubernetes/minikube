package newrelic_platform_go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	NEWRELIC_API_URL = "https://platform-api.newrelic.com/platform/v1/metrics"
)

type INewrelicPlugin interface {
	GetMetricaKey(metrica IMetrica) string
	Harvest() error
	Run()
	AddComponent(component IComponent)
}
type NewrelicPlugin struct {
	Agent      *Agent          `json:"agent"`
	Components []ComponentData `json:"components"`

	ComponentModels      []IComponent `json:"-"`
	LastPollTime         time.Time    `json:"-"`
	Verbose              bool         `json:"-"`
	LicenseKey           string       `json:"-"`
	PollIntervalInSecond int          `json:"-"`
}

func NewNewrelicPlugin(version string, licenseKey string, pollInterval int) *NewrelicPlugin {
	plugin := &NewrelicPlugin{
		LicenseKey:           licenseKey,
		PollIntervalInSecond: pollInterval,
	}

	plugin.Agent = NewAgent(version)
	plugin.Agent.CollectEnvironmentInfo()

	plugin.ComponentModels = []IComponent{}
	return plugin
}

func (plugin *NewrelicPlugin) Harvest() error {
	startTime := time.Now()
	var duration int
	if plugin.LastPollTime.IsZero() {
		duration = plugin.PollIntervalInSecond
	} else {
		duration = int(startTime.Sub(plugin.LastPollTime).Seconds())
	}

	plugin.Components = make([]ComponentData, 0, len(plugin.ComponentModels))
	for i := 0; i < len(plugin.ComponentModels); i++ {
		plugin.ComponentModels[i].SetDuration(duration)
		plugin.Components = append(plugin.Components, plugin.ComponentModels[i].Harvest(plugin))
	}

	if httpCode, err := plugin.SendMetricas(); err != nil {
		log.Printf("Can not send metricas to newrelic: %#v\n", err)
		return err
	} else {

		if plugin.Verbose {
			log.Printf("Got HTTP response code:%d", httpCode)
		}

		if err, isFatal := plugin.CheckResponse(httpCode); isFatal {		
			log.Printf("Got fatal error:%v\n", err)
			return err
		} else {
			if err != nil {
				log.Printf("WARNING: %v", err)
			}
			return err
		}
	}
	return nil
}

func (plugin *NewrelicPlugin) GetMetricaKey(metrica IMetrica) string {
	var keyBuffer bytes.Buffer

	keyBuffer.WriteString("Component/")
	keyBuffer.WriteString(metrica.GetName())
	keyBuffer.WriteString("[")
	keyBuffer.WriteString(metrica.GetUnits())
	keyBuffer.WriteString("]")

	return keyBuffer.String()
}

func (plugin *NewrelicPlugin) SendMetricas() (int, error) {
	client := &http.Client{}
	var metricasJson []byte
	var encodingError error

	if plugin.Verbose {
		metricasJson, encodingError = json.MarshalIndent(plugin, "", "    ")
	} else {
		metricasJson, encodingError = json.Marshal(plugin)
	}

	if encodingError != nil {
		return 0, encodingError
	}

	jsonAsString := string(metricasJson)
	if plugin.Verbose {
		log.Printf("Send data:%s \n", jsonAsString)
	}

	if httpRequest, err := http.NewRequest("POST", NEWRELIC_API_URL, strings.NewReader(jsonAsString)); err != nil {
		return 0, err
	} else {
		httpRequest.Header.Set("X-License-Key", plugin.LicenseKey)
		httpRequest.Header.Set("Content-Type", "application/json")
		httpRequest.Header.Set("Accept", "application/json")

		if httpResponse, err := client.Do(httpRequest); err != nil {
			return 0, err
		} else {
			defer httpResponse.Body.Close()
			return httpResponse.StatusCode, nil
		}
	}

	// we will never get there
	return 0, nil
}

func (plugin *NewrelicPlugin) ClearSentData() {
	for _, component := range plugin.ComponentModels {
		component.ClearSentData()
	}
	plugin.Components = nil
	plugin.LastPollTime = time.Now()
}

func (plugin *NewrelicPlugin) CheckResponse(httpResponseCode int) (error, bool) {
	isFatal := false
	var err error
	switch httpResponseCode {
	case http.StatusOK:
		{
			plugin.ClearSentData()
		}
	case http.StatusForbidden:
		{
			err = fmt.Errorf("Authentication error (no license key header, or invalid license key).\n")
			isFatal = true
		}
	case http.StatusBadRequest:
		{
			err = fmt.Errorf("The request or headers are in the wrong format or the URL is incorrect.\n")
			isFatal = true
		}
	case http.StatusNotFound:
		{
			err = fmt.Errorf("Invalid URL\n")
			isFatal = true
		}
	case http.StatusRequestEntityTooLarge:
		{
			err = fmt.Errorf("Too many metrics were sent in one request, or too many components (instances) were specified in one request, or other single-request limits were reached.\n")
			//discard metrics
			plugin.ClearSentData()
		}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		{
			err = fmt.Errorf("Got %v response code.Metricas will be aggregated", httpResponseCode)
		}
	}
	return err, isFatal
}

func (plugin *NewrelicPlugin) Run() {
	plugin.Harvest()
	tickerChannel := time.Tick(time.Duration(plugin.PollIntervalInSecond) * time.Second)
	for ts := range tickerChannel {
		plugin.Harvest()

		if plugin.Verbose {
			log.Printf("Harvest ended at:%v\n", ts)
		}
	}
}

func (plugin *NewrelicPlugin) AddComponent(component IComponent) {
	plugin.ComponentModels = append(plugin.ComponentModels, component)
}
