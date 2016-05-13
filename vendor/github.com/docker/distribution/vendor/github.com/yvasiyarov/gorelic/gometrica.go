package gorelic

import (
	"fmt"
	metrics "github.com/yvasiyarov/go-metrics"
)

const (
	histogramMin = iota
	histogramMax
	histogramMean
	histogramPercentile
	histogramStdDev
	histogramVariance
	noHistogramFunctions
)

type goMetricaDataSource struct {
	metrics.Registry
}

func (ds goMetricaDataSource) GetGaugeValue(key string) (float64, error) {
	if valueContainer := ds.Get(key); valueContainer == nil {
		return 0, fmt.Errorf("metrica with name %s is not registered\n", key)
	} else if gauge, ok := valueContainer.(metrics.Gauge); ok {
		return float64(gauge.Value()), nil
	} else {
		return 0, fmt.Errorf("metrica container has unexpected type: %T\n", valueContainer)
	}
}

func (ds goMetricaDataSource) GetHistogramValue(key string, statFunction int, percentile float64) (float64, error) {
	if valueContainer := ds.Get(key); valueContainer == nil {
		return 0, fmt.Errorf("metrica with name %s is not registered\n", key)
	} else if histogram, ok := valueContainer.(metrics.Histogram); ok {
		switch statFunction {
		default:
			return 0, fmt.Errorf("unsupported stat function for histogram: %s\n", statFunction)
		case histogramMax:
			return float64(histogram.Max()), nil
		case histogramMin:
			return float64(histogram.Min()), nil
		case histogramMean:
			return float64(histogram.Mean()), nil
		case histogramStdDev:
			return float64(histogram.StdDev()), nil
		case histogramVariance:
			return float64(histogram.Variance()), nil
		case histogramPercentile:
			return float64(histogram.Percentile(percentile)), nil
		}
	} else {
		return 0, fmt.Errorf("metrica container has unexpected type: %T\n", valueContainer)
	}
}

type baseGoMetrica struct {
	dataSource    goMetricaDataSource
	basePath      string
	name          string
	units         string
	dataSourceKey string
}

func (metrica *baseGoMetrica) GetName() string {
	return metrica.basePath + metrica.name
}

func (metrica *baseGoMetrica) GetUnits() string {
	return metrica.units
}

type gaugeMetrica struct {
	*baseGoMetrica
}

func (metrica *gaugeMetrica) GetValue() (float64, error) {
	return metrica.dataSource.GetGaugeValue(metrica.dataSourceKey)
}

type gaugeIncMetrica struct {
	*baseGoMetrica
	previousValue float64
}

func (metrica *gaugeIncMetrica) GetValue() (float64, error) {
	var value float64
	var currentValue float64
	var err error
	if currentValue, err = metrica.dataSource.GetGaugeValue(metrica.dataSourceKey); err == nil {
		value = currentValue - metrica.previousValue
		metrica.previousValue = currentValue
	}
	return value, err
}

type histogramMetrica struct {
	*baseGoMetrica
	statFunction    int
	percentileValue float64
}

func (metrica *histogramMetrica) GetValue() (float64, error) {
	return metrica.dataSource.GetHistogramValue(metrica.dataSourceKey, metrica.statFunction, metrica.percentileValue)
}
