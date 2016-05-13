package gorelic

import (
	metrics "github.com/yvasiyarov/go-metrics"
	"github.com/yvasiyarov/newrelic_platform_go"
	"time"
)

func newMemoryMetricaDataSource(pollInterval int) goMetricaDataSource {
	r := metrics.NewRegistry()

	metrics.RegisterRuntimeMemStats(r)
	metrics.CaptureRuntimeMemStatsOnce(r)
	go metrics.CaptureRuntimeMemStats(r, time.Duration(pollInterval)*time.Second)
	return goMetricaDataSource{r}
}

func addMemoryMericsToComponent(component newrelic_platform_go.IComponent, pollInterval int) {
	gaugeMetrics := []*baseGoMetrica{
		//Memory in use metrics
		&baseGoMetrica{
			name:          "InUse/Total",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.Alloc",
		},
		&baseGoMetrica{
			name:          "InUse/Heap",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.HeapAlloc",
		},
		&baseGoMetrica{
			name:          "InUse/Stack",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.StackInuse",
		},
		&baseGoMetrica{
			name:          "InUse/MSpanInuse",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.MSpanInuse",
		},
		&baseGoMetrica{
			name:          "InUse/MCacheInuse",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.MCacheInuse",
		},
	}
	ds := newMemoryMetricaDataSource(pollInterval)
	for _, m := range gaugeMetrics {
		m.basePath = "Runtime/Memory/"
		m.dataSource = ds
		component.AddMetrica(&gaugeMetrica{m})
	}

	gaugeIncMetrics := []*baseGoMetrica{
		//NO operations graph
		&baseGoMetrica{
			name:          "Operations/NoPointerLookups",
			units:         "lookups",
			dataSourceKey: "runtime.MemStats.Lookups",
		},
		&baseGoMetrica{
			name:          "Operations/NoMallocs",
			units:         "mallocs",
			dataSourceKey: "runtime.MemStats.Mallocs",
		},
		&baseGoMetrica{
			name:          "Operations/NoFrees",
			units:         "frees",
			dataSourceKey: "runtime.MemStats.Frees",
		},

		// Sytem memory allocations
		&baseGoMetrica{
			name:          "SysMem/Total",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.Sys",
		},
		&baseGoMetrica{
			name:          "SysMem/Heap",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.HeapSys",
		},
		&baseGoMetrica{
			name:          "SysMem/Stack",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.StackSys",
		},
		&baseGoMetrica{
			name:          "SysMem/MSpan",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.MSpanSys",
		},
		&baseGoMetrica{
			name:          "SysMem/MCache",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.MCacheSys",
		},
		&baseGoMetrica{
			name:          "SysMem/BuckHash",
			units:         "bytes",
			dataSourceKey: "runtime.MemStats.BuckHashSys",
		},
	}

	for _, m := range gaugeIncMetrics {
		m.basePath = "Runtime/Memory/"
		m.dataSource = ds
		component.AddMetrica(&gaugeIncMetrica{baseGoMetrica: m})
	}
}
