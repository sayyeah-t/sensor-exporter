package sensor

import (
	"sensor-exporter/sensor/bme280"
	"sensor-exporter/sensor/ccs811"
)

type Sensor interface {
	Init(string) error
	GetSensorName() string
	GetMetricsDescriptions() map[string]string
	Update() map[string]float64
	GetConsoleHeader() string
	GetConsoleData() string
	Close()
}

var (
	sensors = []Sensor{}
)

func Init(enabledSensors []string) []Sensor {
	//sensors = make([]Sensor, len(enabledSensors))
	for _, s := range enabledSensors {
		if s == "bme280" {
			sensors = append(sensors, &bme280.BME280{})
		}
		if s == "ccs811" {
			sensors = append(sensors, &ccs811.CCS811{})
		}
	}

	return sensors
}

func GetDescriptions(enabledMetrics []string) map[string]string {
	desc := make(map[string]string, len(enabledMetrics))
	for _, metrics := range enabledMetrics {
		for _, s := range sensors {
			for i, metricsDescription := range s.GetMetricsDescriptions() {
				if metrics == i {
					desc[metrics] = metricsDescription
				}
			}
		}
	}

	return desc
}
