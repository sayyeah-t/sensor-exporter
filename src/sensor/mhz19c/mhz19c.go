package mhz19c

import (
	"fmt"
	"log"
	"math"
	"sensor-exporter/config"
	"time"

	"github.com/tarm/serial"
)

var (
	conf config.Mhz19c

	// data for writing to get co2 data (from data sheet)
	read_co2_data = []byte{
		0xFF,
		0x01,
		0x86,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x79,
	}
	// data for writing to enable zero point calibration (from data sheet)
	// last byte is checksum: 0xFF - ([1]+[2]+...+[7]) + 0x01
	zero_point_calibration_on = []byte{
		0xFF,
		0x01,
		0x79,
		0xA0,
		0x00,
		0x00,
		0x00,
		0x00,
		0xE6,
	}
	// data for writing to disable zero point calibration (from data sheet)
	// last byte is checksum: 0xFF - ([1]+[2]+...+[7]) + 0x01
	zero_point_calibration_off = []byte{
		0xFF,
		0x01,
		0x79,
		0x00,
		0x00,
		0x00,
		0x00,
		0x00,
		0x86,
	}
)

type MHZ19C struct {
	data map[string]float64
	port *serial.Port
}

func (m *MHZ19C) Init() error {
	conf = config.GetConfig().Mhz19c
	log.Println("Open sensor MH-Z19C")

	// setup serial port
	c := &serial.Config{Name: conf.SerialPort, Baud: conf.SerialBaudrate, ReadTimeout: time.Millisecond * 500}
	port, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	m.port = port

	// set auto calibration
	if conf.SelfCalibration {
		m.port.Write(zero_point_calibration_on)
	} else {
		m.port.Write(zero_point_calibration_off)
	}

	// read existing data before start application
	buf := make([]byte, 9)
	flag := true
	for flag {
		count, _ := m.port.Read(buf)
		if count == 0 {
			flag = false
		}
		time.Sleep(time.Second * 1)
	}

	// init data array
	m.data = map[string]float64{
		conf.Co2MetricsName: 0.0,
	}

	return nil
}

func (m *MHZ19C) Close() {
	log.Println("Close sensor MH-Z19C")
	m.port.Close()
}

func (m *MHZ19C) GetSensorName() string {
	return "MH-Z19C"
}

func (m *MHZ19C) GetMetricsDescriptions() map[string]string {
	return map[string]string{
		conf.Co2MetricsName: "CO2 value in [ppm] measured by MH-Z19C",
	}
}

func (m *MHZ19C) Update() map[string]float64 {
	buf := make([]byte, 9)
	_, werr := m.port.Write(read_co2_data)
	_, rerr := m.port.Read(buf)
	if werr != nil || rerr != nil {
		log.Printf("MH-Z19C data update error")
		return m.data
	}
	value := int(buf[2])<<8 | int(buf[3])
	m.data[conf.Co2MetricsName] = math.Min(10000.0, math.Max(400.0, float64(value)))

	return m.data
}

func (m *MHZ19C) GetConsoleHeader() string {
	return " CO2[ppm] "
}

func (m *MHZ19C) GetConsoleData() string {
	msg := fmt.Sprintf(" %8.2f ", m.data[conf.Co2MetricsName])
	return msg
}
