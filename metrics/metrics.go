package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

type dummyGauge struct {
	prometheus.Gauge
}

func (d *dummyGauge) Inc() {}
func (d *dummyGauge) Dec() {}

type dummyHistogram struct {
	prometheus.Histogram
}

const defaultNamespace string = "eventinator"

func (d *dummyHistogram) Observe(_ float64) {}

func constructHistogram(enabled bool, name string, help string) prometheus.Histogram {
	if enabled {
		histogram := prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:      name,
				Namespace: defaultNamespace,
				Help:      help,
			},
		)
		prometheus.MustRegister(histogram)
		return histogram
	}
	return &dummyHistogram{}
}

func constructGauge(enabled bool, name string, help string) prometheus.Gauge {
	if enabled {
		gauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      name,
				Namespace: defaultNamespace,
				Help:      help,
			},
		)
		prometheus.MustRegister(gauge)
		return gauge
	}
	return &dummyGauge{}
}

type Instrumentation struct {
	Enabled               bool
	Publishers            prometheus.Gauge
	MessagesReceived      prometheus.Gauge
	Subscribers           prometheus.Gauge
	MessagesDelivered     prometheus.Gauge
	PublisherPerformance  prometheus.Histogram
	SubscriberPerformance prometheus.Histogram
}

func NewInstrumentation(enabled bool) *Instrumentation {
	return &Instrumentation{
		Enabled:               enabled,
		Publishers:            constructGauge(enabled, "publishers", "Number of active publishers"),
		MessagesReceived:      constructGauge(enabled, "messages_received", "Number of messages received"),
		Subscribers:           constructGauge(enabled, "subscribers", "Number of active subscribers"),
		MessagesDelivered:     constructGauge(enabled, "messages_delivered", "Number of delivered messages"),
		PublisherPerformance:  constructHistogram(enabled, "publisher_perf", "Publish performance"),
		SubscriberPerformance: constructHistogram(enabled, "subscriber_perf", "Subscribe performance"),
	}
}

func (i *Instrumentation) Serve(port int) {
	go func() {
		if i.Enabled {
			http.Handle("/metrics", promhttp.Handler())
			bindAddr := fmt.Sprintf(":%d", port)
			log.Infof("Prometheus metrics can be scraped from this host on port %d.", port)
			log.Error(http.ListenAndServe(bindAddr, nil))
		} else {
			log.Info("Metrics scraping disabled.")
		}
	}()
}
