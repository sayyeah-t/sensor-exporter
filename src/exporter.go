package main

import (
	"context"
	"log"
	"net/http"
	"sensor-exporter/sensor"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	srv       *http.Server
	reg       = prometheus.NewRegistry()
	gaugeVecs map[string]*prometheus.GaugeVec
)

func initExporter() {
	desc := sensor.GetDescriptions(conf.ExportMetrics)
	gaugeVecs = make(map[string]*prometheus.GaugeVec, len(conf.ExportMetrics))
	for _, metrics := range conf.ExportMetrics {
		gaugeVecs[metrics] = promauto.With(reg).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: metrics,
				Help: desc[metrics],
			},
			[]string{"sensor_name"},
		)
	}

}

func setExportValue(metricsName string, sensorName string) {
	gaugeVecs[metricsName].WithLabelValues(sensorName).Set(data[metricsName])
}

func runExporter() {
	srv = &http.Server{Addr: conf.BindIp + ":" + conf.BindPort}
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(srv.ListenAndServe())
}

func stopExporter() error {
	if err := srv.Shutdown(context.Background()); err != nil {
		return err
	}

	return nil
}
