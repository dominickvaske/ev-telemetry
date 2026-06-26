package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	EventsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ev",
			Name:      "events_counter_total",
			Help:      "Total number of events processed since start",
		},
	)

	AlertsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ev",
			Name:      "alerts_counter_total",
			Help:      "total number of alerts that have come in since start",
		},
	)

	RoutineGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ev",
			Name:      "go_routines_gauge",
			Help:      "number of go routines currently running",
		},
	)

	LatencyHist = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "ev",
			Name:      "latency_histogram",
			Help:      "histogram of all the latency times experienced in processing",
		},
	)
)

func init() {
	prometheus.MustRegister(EventsCounter)
	prometheus.MustRegister(AlertsCounter)
	prometheus.MustRegister(RoutineGauge)
	prometheus.MustRegister(LatencyHist)
}
