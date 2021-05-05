package config

import (
	"log"

	"sensor-exporter/util"

	"gopkg.in/ini.v1"
)

type Default struct {
	BindIp         string
	BindPort       string
	I2cDevice      string
	EnabledSensors []string
	ExportMetrics  []string
}

type Bme280 struct {
	Address                int
	TemperatureMetricsName string
	HumidityMetricsName    string
	PressureMetricsName    string
}

type Ccs811 struct {
	Address        int
	Co2MetricsName string
	VocMetricsName string
}

type Config struct {
	Default Default
	Bme280  Bme280
	Ccs811  Ccs811
}

var (
	configuration Config
)

func Init(configPath string) error {
	cfg, err := ini.InsensitiveLoad(configPath)
	if err != nil {
		return err
	}
	configuration = Config{
		Default: Default{
			BindIp:         cfg.Section("default").Key("bind_ip").MustString("0.0.0.0"),
			BindPort:       cfg.Section("default").Key("bind_port").MustString("8080"),
			I2cDevice:      cfg.Section("default").Key("i2c_device").MustString("/dev/i2c-1"),
			EnabledSensors: util.ParseStringToSlice(cfg.Section("default").Key("enable_sensor").MustString("bme280,ccs811")),
			ExportMetrics:  util.ParseStringToSlice(cfg.Section("default").Key("export_metrics").MustString("temperature,humidity,pressure,co2,voc")),
		},
		Bme280: Bme280{
			Address:                cfg.Section("bme280").Key("i2c_address").MustInt(0x76),
			TemperatureMetricsName: cfg.Section("bme280").Key("metrics_name_temp").MustString("temperature"),
			HumidityMetricsName:    cfg.Section("bme280").Key("metrics_name_humid").MustString("humidity"),
			PressureMetricsName:    cfg.Section("bme280").Key("metrics_name_press").MustString("pressure"),
		},
		Ccs811: Ccs811{
			Address:        cfg.Section("ccs811").Key("i2c_address").MustInt(0x5a),
			Co2MetricsName: cfg.Section("ccs811").Key("metrics_name_co2").MustString("co2"),
			VocMetricsName: cfg.Section("ccs811").Key("metrics_name_voc").MustString("voc"),
		},
	}
	return nil
}

func GetConfig() Config {
	return configuration
}

func DumpConfig() {
	log.Printf("Condig dump:\n%+v\n", configuration)
}
