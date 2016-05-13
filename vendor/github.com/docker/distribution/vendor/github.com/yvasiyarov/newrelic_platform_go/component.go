package newrelic_platform_go

import (
	"log"
	"math"
)

type ComponentData interface{}
type IComponent interface {
	Harvest(plugin INewrelicPlugin) ComponentData
	SetDuration(duration int)
	AddMetrica(model IMetrica)
	ClearSentData()
}

type PluginComponent struct {
	Name          string                  `json:"name"`
	GUID          string                  `json:"guid"`
	Duration      int                     `json:"duration"`
	Metrics       map[string]MetricaValue `json:"metrics"`
	MetricaModels []IMetrica              `json:"-"`
}

func NewPluginComponent(name string, guid string) *PluginComponent {
	c := &PluginComponent{
		Name: name,
		GUID: guid,
	}
	return c
}

func (component *PluginComponent) AddMetrica(model IMetrica) {
	component.MetricaModels = append(component.MetricaModels, model)
}

func (component *PluginComponent) ClearSentData() {
	component.Metrics = nil
}

func (component *PluginComponent) SetDuration(duration int) {
	component.Duration = duration
}

func (component *PluginComponent) Harvest(plugin INewrelicPlugin) ComponentData {
	component.Metrics = make(map[string]MetricaValue, len(component.MetricaModels))
	for i := 0; i < len(component.MetricaModels); i++ {
		model := component.MetricaModels[i]
		metricaKey := plugin.GetMetricaKey(model)

		if newValue, err := model.GetValue(); err == nil {
		        if math.IsInf(newValue, 0) || math.IsNaN(newValue) {
                                newValue = 0
                        }

			if existMetric, ok := component.Metrics[metricaKey]; ok {
				if floatExistVal, ok := existMetric.(float64); ok {
					component.Metrics[metricaKey] = NewAggregatedMetricaValue(floatExistVal, newValue)
				} else if aggregatedValue, ok := existMetric.(*AggregatedMetricaValue); ok {
					aggregatedValue.Aggregate(newValue)
				} else {
					panic("Invalid type in metrica value")
				}
			} else {
				component.Metrics[metricaKey] = newValue
			}
		} else {
			log.Printf("Can not get metrica: %v, got error:%#v", model.GetName(), err)
		}
	}
	return component
}
