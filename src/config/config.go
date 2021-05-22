package config

import (
	"log"

	"sensor-exporter/util"

	"gopkg.in/ini.v1"
)

type Default struct {
	BindIp         string
	BindPort       string
	EnabledSensors []string
	ExportMetrics  []string
}

type Bme280 struct {
	I2cDevice              string
	I2cAddress             int
	TemperatureMetricsName string
	HumidityMetricsName    string
	PressureMetricsName    string
}

type Ccs811 struct {
	I2cDevice      string
	I2cAddress     int
	Co2MetricsName string
	VocMetricsName string
	Baseline       int
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
			EnabledSensors: util.ParseStringToSlice(cfg.Section("default").Key("enable_sensor").MustString("bme280,ccs811")),
			ExportMetrics:  util.ParseStringToSlice(cfg.Section("default").Key("export_metrics").MustString("temperature,humidity,pressure,co2,voc")),
		},
		Bme280: Bme280{
			I2cDevice:              cfg.Section("bme280").Key("i2c_device").MustString("/dev/i2c-1"),
			I2cAddress:             cfg.Section("bme280").Key("i2c_address").MustInt(0x76),
			TemperatureMetricsName: cfg.Section("bme280").Key("metrics_name_temp").MustString("temperature"),
			HumidityMetricsName:    cfg.Section("bme280").Key("metrics_name_humid").MustString("humidity"),
			PressureMetricsName:    cfg.Section("bme280").Key("metrics_name_press").MustString("pressure"),
		},
		Ccs811: Ccs811{
			I2cDevice:      cfg.Section("ccs811").Key("i2c_device").MustString("/dev/i2c-1"),
			I2cAddress:     cfg.Section("ccs811").Key("i2c_address").MustInt(0x5a),
			Co2MetricsName: cfg.Section("ccs811").Key("metrics_name_co2").MustString("co2"),
			VocMetricsName: cfg.Section("ccs811").Key("metrics_name_voc").MustString("voc"),
			Baseline:       cfg.Section("ccs811").Key("baseline").MustInt(0),
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
