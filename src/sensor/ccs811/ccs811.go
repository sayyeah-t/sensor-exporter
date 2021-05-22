package ccs811

import (
	"errors"
	"fmt"
	"log"
	"sensor-exporter/config"
	"time"

	"golang.org/x/exp/io/i2c"
)

var (
	conf config.Ccs811

	// Addresses
	status          = byte(0x00)
	meas_mode       = byte(0x01)
	alg_result_data = byte(0x02)
	//raw_data        = byte(0x03)
	//env_data        = byte(0x05)
	//ntc             = byte(0x06)
	//thresholds      = byte(0x10)
	baseline = byte(0x11)
	hw_id    = byte(0x20)
	//hw_version      = byte(0x21)
	//fw_boot_version = byte(0x23)
	//fw_app_version  = byte(0x24)
	error_id  = byte(0xE0)
	app_start = byte(0xF4)
	//sw_reset        = byte(0xFF)

	// Config values
	mode = byte(0x01) // 1 sec drive mode

)

type CCS811 struct {
	data          map[string]float64
	dev           *i2c.Device
	baseline      uint16
	baselineCount int
	wakeupFlag    bool
}

func (c *CCS811) Init() error {
	// init vars
	conf = config.GetConfig().Ccs811
	log.Println("Open sensor CCS811")

	c.data = map[string]float64{
		conf.Co2MetricsName: 0.0,
		conf.VocMetricsName: 0.0,
	}

	dev, err := i2c.Open(&i2c.Devfs{Dev: conf.I2cDevice}, conf.I2cAddress)
	if err != nil {
		return err
	}
	c.dev = dev

	c.baselineCount = 0
	c.baseline = uint16(conf.Baseline)
	c.wakeupFlag = false
	if c.baseline == 0 {
		c.wakeupFlag = true
	}

	// validate ccs811
	//// hardware id
	val_hw_id := make([]byte, 1)
	if err := c.dev.ReadReg(hw_id, val_hw_id); err != nil {
		return err
	}
	if val_hw_id[0] != 0x81 {
		return errors.New("hardware id wasn't match 0x81")
	}
	//// check device error
	if err := c.checkError(); err != nil {
		return err
	}
	//// validate application
	if err := c.validateApplication(); err != nil {
		return err
	}

	// app start
	//// write empty data to start app
	if err := c.dev.WriteReg(app_start, nil); err != nil {
		return err
	}
	//// check device error
	if err := c.checkError(); err != nil {
		return err
	}
	//// set drive mode
	if err := c.setDriveMode(); err != nil {
		return err
	}
	//// check device error
	if err := c.checkError(); err != nil {
		return err
	}

	// waiting for start sensor
	for {
		c.Update()
		time.Sleep(1 * time.Second)
		if c.data[conf.Co2MetricsName] > 0 {
			// min co2 value is 400 if the sensor is running
			break
		}
	}

	return nil
}

func (c *CCS811) getBaseline() uint16 {
	val_baseline := make([]byte, 2)
	if err := c.dev.ReadReg(baseline, val_baseline); err != nil {
		return 0
	}
	log.Printf("Check baseline: %d", uint16(val_baseline[0])<<8|uint16(val_baseline[1]))

	return uint16(val_baseline[0])<<8 | uint16(val_baseline[1])
}

func (c *CCS811) setBaseline() uint16 {
	val_baseline := []byte{
		byte((c.baseline >> 8) & 0xFF),
		byte(c.baseline & 0xFF),
	}
	c.dev.WriteReg(baseline, val_baseline)

	return c.getBaseline()
}

func (c *CCS811) checkError() error {
	errorMsg := "error: "
	val_hw_status := make([]byte, 1)
	if err := c.dev.ReadReg(status, val_hw_status); err != nil {
		return err
	}
	if (val_hw_status[0] & (1 << 0)) == 1 {
		val_hw_error := make([]byte, 1)
		if err := c.dev.ReadReg(error_id, val_hw_error); err != nil {
			return err
		}
		if (val_hw_error[0] & (1 << 5)) != 0 {
			errorMsg = errorMsg + "heater supply"
		} else if (val_hw_error[0] & (1 << 4)) != 0 {
			errorMsg = errorMsg + "heater fault"
		} else if (val_hw_error[0] & (1 << 3)) != 0 {
			errorMsg = errorMsg + "max resistance"
		} else if (val_hw_error[0] & (1 << 2)) != 0 {
			errorMsg = errorMsg + "meas mode invalid"
		} else if (val_hw_error[0] & (1 << 1)) != 0 {
			errorMsg = errorMsg + "read register invalid"
		} else if (val_hw_error[0] & (1 << 0)) != 0 {
			errorMsg = errorMsg + "message invalid"
		}
		return errors.New(errorMsg)
	}

	return nil
}

func (c *CCS811) validateApplication() error {
	errorMsg := "error: "
	val_hw_status := make([]byte, 1)
	if err := c.dev.ReadReg(status, val_hw_status); err != nil {
		return err
	}
	if (val_hw_status[0] & (1 << 4)) != 16 {
		return errors.New(errorMsg + "validation application error")
	}

	return nil
}

func (c *CCS811) setDriveMode() error {
	setting := make([]byte, 1)
	if err := c.dev.ReadReg(meas_mode, setting); err != nil {
		return err
	}
	setting[0] = setting[0] & (^(byte(7) << 4))
	setting[0] = setting[0] | (mode << 4)
	if err := c.dev.WriteReg(meas_mode, setting); err != nil {
		return err
	}

	return nil
}

func (c *CCS811) Close() {
	log.Println("Close sensor CCS811")
	c.dev.Close()
}

func (c *CCS811) GetSensorName() string {
	return "CCS811"
}

func (c *CCS811) GetMetricsDescriptions() map[string]string {
	return map[string]string{
		conf.Co2MetricsName: "CO2 value in [ppm] measured by CCS811",
		conf.VocMetricsName: "VOC value in [ppb] measured by CCS811",
	}
}

func (c *CCS811) Update() map[string]float64 {
	c.baselineCount = c.baselineCount + 1
	if c.baselineCount%1200 == 0 {
		c.baselineCount = 0
		if c.wakeupFlag {
			c.baseline = c.getBaseline()
			c.wakeupFlag = false
		}
		c.setBaseline()
	}
	data_available := make([]byte, 1)
	if err := c.dev.ReadReg(status, data_available); err != nil {
		log.Println("device is not available")
		return c.data
	}
	if (data_available[0] & (1 << 3)) > 0 {
		result_data := make([]byte, 4)
		if err := c.dev.ReadReg(alg_result_data, result_data); err != nil {
			log.Println("ccs811 read data error")
			return c.data
		}
		c.data[conf.Co2MetricsName] = float64((int16(result_data[0]) << 8) | int16(result_data[1]))
		c.data[conf.VocMetricsName] = float64((int16(result_data[2]) << 8) | int16(result_data[3]))
	}

	return c.data
}

func (c *CCS811) GetConsoleHeader() string {
	return " CO2[ppm] | VOC[ppb] "
}

func (c *CCS811) GetConsoleData() string {
	msg := fmt.Sprintf(" %8.2f | %8.2f ", c.data[conf.Co2MetricsName], c.data[conf.VocMetricsName])
	return msg
}
