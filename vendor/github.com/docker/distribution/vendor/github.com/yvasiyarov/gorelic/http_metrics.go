package gorelic

import (
	metrics "github.com/yvasiyarov/go-metrics"
	"github.com/yvasiyarov/newrelic_platform_go"
	"net/http"
	"time"
)

type tHTTPHandlerFunc func(http.ResponseWriter, *http.Request)
type tHTTPHandler struct {
	originalHandler     http.Handler
	originalHandlerFunc tHTTPHandlerFunc
	isFunc              bool
	timer               metrics.Timer
}

var httpTimer metrics.Timer

func newHTTPHandlerFunc(h tHTTPHandlerFunc) *tHTTPHandler {
	return &tHTTPHandler{
		isFunc:              true,
		originalHandlerFunc: h,
	}
}
func newHTTPHandler(h http.Handler) *tHTTPHandler {
	return &tHTTPHandler{
		isFunc:          false,
		originalHandler: h,
	}
}

func (handler *tHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()
	defer handler.timer.UpdateSince(startTime)

	if handler.isFunc {
		handler.originalHandlerFunc(w, req)
	} else {
		handler.originalHandler.ServeHTTP(w, req)
	}
}

type baseTimerMetrica struct {
	dataSource metrics.Timer
	name       string
	units      string
}

func (metrica *baseTimerMetrica) GetName() string {
	return metrica.name
}

func (metrica *baseTimerMetrica) GetUnits() string {
	return metrica.units
}

type timerRate1Metrica struct {
	*baseTimerMetrica
}

func (metrica *timerRate1Metrica) GetValue() (float64, error) {
	return metrica.dataSource.Rate1(), nil
}

type timerRateMeanMetrica struct {
	*baseTimerMetrica
}

func (metrica *timerRateMeanMetrica) GetValue() (float64, error) {
	return metrica.dataSource.RateMean(), nil
}

type timerMeanMetrica struct {
	*baseTimerMetrica
}

func (metrica *timerMeanMetrica) GetValue() (float64, error) {
	return metrica.dataSource.Mean() / float64(time.Millisecond), nil
}

type timerMinMetrica struct {
	*baseTimerMetrica
}

func (metrica *timerMinMetrica) GetValue() (float64, error) {
	return float64(metrica.dataSource.Min()) / float64(time.Millisecond), nil
}

type timerMaxMetrica struct {
	*baseTimerMetrica
}

func (metrica *timerMaxMetrica) GetValue() (float64, error) {
	return float64(metrica.dataSource.Max()) / float64(time.Millisecond), nil
}

type timerPercentile75Metrica struct {
	*baseTimerMetrica
}

func (metrica *timerPercentile75Metrica) GetValue() (float64, error) {
	return metrica.dataSource.Percentile(0.75) / float64(time.Millisecond), nil
}

type timerPercentile90Metrica struct {
	*baseTimerMetrica
}

func (metrica *timerPercentile90Metrica) GetValue() (float64, error) {
	return metrica.dataSource.Percentile(0.90) / float64(time.Millisecond), nil
}

type timerPercentile95Metrica struct {
	*baseTimerMetrica
}

func (metrica *timerPercentile95Metrica) GetValue() (float64, error) {
	return metrica.dataSource.Percentile(0.95) / float64(time.Millisecond), nil
}

func addHTTPMericsToComponent(component newrelic_platform_go.IComponent, timer metrics.Timer) {
	rate1 := &timerRate1Metrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/throughput/1minute",
			units:      "rps",
			dataSource: timer,
		},
	}
	component.AddMetrica(rate1)

	rateMean := &timerRateMeanMetrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/throughput/rateMean",
			units:      "rps",
			dataSource: timer,
		},
	}
	component.AddMetrica(rateMean)

	responseTimeMean := &timerMeanMetrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/responseTime/mean",
			units:      "ms",
			dataSource: timer,
		},
	}
	component.AddMetrica(responseTimeMean)

	responseTimeMax := &timerMaxMetrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/responseTime/max",
			units:      "ms",
			dataSource: timer,
		},
	}
	component.AddMetrica(responseTimeMax)

	responseTimeMin := &timerMinMetrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/responseTime/min",
			units:      "ms",
			dataSource: timer,
		},
	}
	component.AddMetrica(responseTimeMin)

	responseTimePercentile75 := &timerPercentile75Metrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/responseTime/percentile75",
			units:      "ms",
			dataSource: timer,
		},
	}
	component.AddMetrica(responseTimePercentile75)

	responseTimePercentile90 := &timerPercentile90Metrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/responseTime/percentile90",
			units:      "ms",
			dataSource: timer,
		},
	}
	component.AddMetrica(responseTimePercentile90)

	responseTimePercentile95 := &timerPercentile95Metrica{
		baseTimerMetrica: &baseTimerMetrica{
			name:       "http/responseTime/percentile95",
			units:      "ms",
			dataSource: timer,
		},
	}
	component.AddMetrica(responseTimePercentile95)
}
