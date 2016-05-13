package newrelic_platform_go

import (
	"math"
)

type IMetrica interface {
	GetValue() (float64, error)
	GetName() string
	GetUnits() string
}

type MetricaValue interface{}

type SimpleMetricaValue float64

type AggregatedMetricaValue struct {
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Total        float64 `json:"total"`
	Count        int     `json:"count"`
	SumOfSquares float64 `json:"sum_of_squares"`
}

func NewAggregatedMetricaValue(existValue float64, newValue float64) *AggregatedMetricaValue {
	v := &AggregatedMetricaValue{
		Min:          math.Min(newValue, existValue),
		Max:          math.Max(newValue, existValue),
		Total:        newValue + existValue,
		Count:        2,
		SumOfSquares: newValue*newValue + existValue*existValue,
	}
	return v
}

func (aggregatedValue *AggregatedMetricaValue) Aggregate(newValue float64) {
	aggregatedValue.Min = math.Min(newValue, aggregatedValue.Min)
	aggregatedValue.Max = math.Max(newValue, aggregatedValue.Max)
	aggregatedValue.Total += newValue
	aggregatedValue.Count++
	aggregatedValue.SumOfSquares += newValue * newValue
}
