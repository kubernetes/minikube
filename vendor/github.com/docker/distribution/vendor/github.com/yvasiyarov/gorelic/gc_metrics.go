package gorelic

import (
	metrics "github.com/yvasiyarov/go-metrics"
	"github.com/yvasiyarov/newrelic_platform_go"
	"time"
)

func newGCMetricaDataSource(pollInterval int) goMetricaDataSource {
	r := metrics.NewRegistry()

	metrics.RegisterDebugGCStats(r)
	go metrics.CaptureDebugGCStats(r, time.Duration(pollInterval)*time.Second)
	return goMetricaDataSource{r}
}

func addGCMericsToComponent(component newrelic_platform_go.IComponent, pollInterval int) {
	metrics := []*baseGoMetrica{
		&baseGoMetrica{
			name:          "NumberOfGCCalls",
			units:         "calls",
			dataSourceKey: "debug.GCStats.NumGC",
		},
		&baseGoMetrica{
			name:          "PauseTotalTime",
			units:         "nanoseconds",
			dataSourceKey: "debug.GCStats.PauseTotal",
		},
	}

	ds := newGCMetricaDataSource(pollInterval)
	for _, m := range metrics {
		m.basePath = "Runtime/GC/"
		m.dataSource = ds
		component.AddMetrica(&gaugeMetrica{m})
	}

	histogramMetrics := []*histogramMetrica{
		&histogramMetrica{
			statFunction:  histogramMax,
			baseGoMetrica: &baseGoMetrica{name: "Max"},
		},
		&histogramMetrica{
			statFunction:  histogramMin,
			baseGoMetrica: &baseGoMetrica{name: "Min"},
		},
		&histogramMetrica{
			statFunction:  histogramMean,
			baseGoMetrica: &baseGoMetrica{name: "Mean"},
		},
		&histogramMetrica{
			statFunction:    histogramPercentile,
			percentileValue: 0.95,
			baseGoMetrica:   &baseGoMetrica{name: "Percentile95"},
		},
	}
	for _, m := range histogramMetrics {
		m.baseGoMetrica.units = "nanoseconds"
		m.baseGoMetrica.dataSourceKey = "debug.GCStats.Pause"
		m.baseGoMetrica.basePath = "Runtime/GC/GCTime/"
		m.baseGoMetrica.dataSource = ds

		component.AddMetrica(m)
	}
}
