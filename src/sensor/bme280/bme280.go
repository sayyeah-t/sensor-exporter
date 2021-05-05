package bme280

import (
	"fmt"
	"log"
	"sensor-exporter/config"

	"golang.org/x/exp/io/i2c"
)

var (
	conf config.Bme280

	// Register ctrl_hum (addr: 0xF2)
	reg_ctrl_hum = byte(0xF2)
	osrs_h       = 1            // Humidity oversampling (3 bits)
	num_ctrl_hum = byte(osrs_h) // write 00000001 to 0xF2

	// Register ctrl_meas (addr: 0xF4)
	reg_ctrl_meas = byte(0xF4)
	osrs_t        = 1                                          // Temperature oversampling (3 bits)
	osrs_p        = 1                                          // Pressure oversampling (3 bits)
	mode          = 3                                          // Normal mode (2 bits)
	num_ctrl_meas = byte((osrs_t << 5) | (osrs_p << 2) | mode) // 3, 3, 2

	// Register config (addr: 0xF5)
	reg_config = byte(0xF5)
	t_sb       = 5                                            // Tstandby 1000ms (3 bits)
	filter     = 3                                            // Filter off (3 bits)
	spi3w_en   = 0                                            // Disable spi3w (1 bit)
	num_config = byte((t_sb << 5) | (filter << 2) | spi3w_en) // 3, 3, 1(not used), 1

	// Addresses of calibration data
	calib_addr1 = byte(0x88) // 24 bytes from here
	calib_addr2 = byte(0xA1) // 1 byte from here
	calib_addr3 = byte(0xE1) // 7 bytes from here

	// Calibration data
	calib_temp1  uint16
	calib_temp2  int16
	calib_temp3  int16
	calib_humid1 uint8
	calib_humid2 int16
	calib_humid3 uint8
	calib_humid4 int16
	calib_humid5 int16
	calib_humid6 int8
	calib_press1 uint16
	calib_press2 int16
	calib_press3 int16
	calib_press4 int16
	calib_press5 int16
	calib_press6 int16
	calib_press7 int16
	calib_press8 int16
	calib_press9 int16
	t_fine       int

	// Addresses of measured value
	temp_msb = byte(0xFA)
	//temp_lsb  = byte(0xFB)
	hum_msb = byte(0xFD)
	//hum_lsb   = byte(0xFE)
	press_msb = byte(0xF7)
	//press_lsb = byte(0xF8)

	bufTemp  = make([]byte, 3)
	bufHumid = make([]byte, 2)
	bufPress = make([]byte, 3)
)

type BME280 struct {
	data map[string]float64
	dev  *i2c.Device
}

func (b *BME280) Init(address string) error {
	conf = config.GetConfig().Bme280
	log.Println("Open sensor BME280")

	b.data = map[string]float64{
		conf.TemperatureMetricsName: 0.0,
		conf.HumidityMetricsName:    0.0,
		conf.PressureMetricsName:    0.0,
	}

	dev, err := i2c.Open(&i2c.Devfs{Dev: address}, conf.Address)
	if err != nil {
		return err
	}
	b.dev = dev
	if err := b.dev.WriteReg(reg_ctrl_hum, []byte{num_ctrl_hum}); err != nil {
		return err
	}
	if err := b.dev.WriteReg(reg_ctrl_meas, []byte{num_ctrl_meas}); err != nil {
		return err
	}
	if err := b.dev.WriteReg(reg_config, []byte{num_config}); err != nil {
		return err
	}
	if err := b.InitCalibrationData(); err != nil {
		return err
	}

	return nil
}

func (b *BME280) InitCalibrationData() error {
	tmpdata24 := make([]byte, 24)
	tmpdata1 := make([]byte, 1)
	tmpdata7 := make([]byte, 7)
	if err := b.dev.ReadReg(calib_addr1, tmpdata24); err != nil {
		return err
	}
	if err := b.dev.ReadReg(calib_addr2, tmpdata1); err != nil {
		return err
	}
	if err := b.dev.ReadReg(calib_addr3, tmpdata7); err != nil {
		return err
	}
	calib_temp1 = (uint16(tmpdata24[1]) << 8) | uint16(tmpdata24[0])
	calib_temp2 = (int16(tmpdata24[3]) << 8) | int16(tmpdata24[2])
	calib_temp3 = (int16(tmpdata24[5]) << 8) | int16(tmpdata24[4])
	calib_press1 = (uint16(tmpdata24[7]) << 8) | uint16(tmpdata24[6])
	calib_press2 = (int16(tmpdata24[9]) << 8) | int16(tmpdata24[8])
	calib_press3 = (int16(tmpdata24[11]) << 8) | int16(tmpdata24[10])
	calib_press4 = (int16(tmpdata24[13]) << 8) | int16(tmpdata24[12])
	calib_press5 = (int16(tmpdata24[15]) << 8) | int16(tmpdata24[14])
	calib_press6 = (int16(tmpdata24[17]) << 8) | int16(tmpdata24[16])
	calib_press7 = (int16(tmpdata24[19]) << 8) | int16(tmpdata24[18])
	calib_press8 = (int16(tmpdata24[21]) << 8) | int16(tmpdata24[20])
	calib_press9 = (int16(tmpdata24[23]) << 8) | int16(tmpdata24[22])
	calib_humid1 = tmpdata1[0]
	calib_humid2 = (int16(tmpdata7[1]) << 8) | int16(tmpdata7[0])
	calib_humid3 = tmpdata7[2]
	calib_humid4 = (int16(tmpdata7[3]) << 4) | (0x0F & int16(tmpdata7[4]))
	calib_humid5 = (int16(tmpdata7[5]) << 4) | ((int16(tmpdata7[4]) >> 4) & 0x0F)
	calib_humid6 = int8(tmpdata7[6])

	return nil
}

func (b *BME280) Close() {
	log.Println("Close sensor BME280")
	b.dev.Close()
}

func (b *BME280) GetSensorName() string {
	return "BME280"
}

func (b *BME280) GetMetricsDescriptions() map[string]string {
	return map[string]string{
		conf.TemperatureMetricsName: "Temperature value in [°C] measured by BME280",
		conf.HumidityMetricsName:    "Humidity value in [%] measured by BME280",
		conf.PressureMetricsName:    "Pressure value in [hPa] measured by BME280",
	}
}

func (b *BME280) calibrateTemp(rawValue int64) float64 {
	var1 := (float64(rawValue)/16384.0 - float64(calib_temp1)/1024.0) * float64(calib_temp2)
	var2 := (float64(rawValue)/131072.0 - float64(calib_temp1)/8192.0)
	var2 = var2 * var2 * float64(calib_temp3)
	t_fine = int(var1 + var2)
	temp := (var1 + var2) / 5120.0
	if temp < -40 {
		temp = -40
	}
	if temp > 85 {
		temp = 85
	}
	return temp
}

func (b *BME280) calibrateHumid(rawValue int64) float64 {
	var1 := float64(t_fine) - 76800.0
	var2 := float64(calib_humid4)*64.0 + (float64(calib_humid5)/16384.0)*var1
	var3 := float64(rawValue) - var2
	var4 := float64(calib_humid2) / 65536.0
	var5 := 1.0 + (float64(calib_humid3)/67108864.0)*var1
	var6 := 1.0 + float64(calib_humid6)/67108864.0*var1*var5
	var6 = var3 * var4 * (var5 * var6)
	humid := var6 * (1.0 - float64(calib_humid1)*var6/524288.0)
	if humid < 0.0 {
		humid = 0.0
	}
	if humid > 100.0 {
		humid = 100.0
	}

	return humid
}

func (b *BME280) calibratePress(rawValue int64) float64 {
	var1 := float64(t_fine)/2.0 - 64000.0
	var2 := var1 * var1 * float64(calib_press6) / 32768.0
	var2 = var2 + var1*float64(calib_press5)*2.0
	var2 = var2/4.0 + float64(calib_press4)*65536.0
	var3 := float64(calib_press3) * var1 * var1 / 524288.0
	var1 = (var3 + float64(calib_press2)*var1) / 524288.0
	var1 = (1.0 + var1/32768.0) * float64(calib_press1)
	var pressure float64 = 30000.0 // min
	if var1 > 0.0 {
		pressure = 1048576.0 - float64(rawValue)
		pressure = (pressure - var2/4096.0) * 6250.0 / var1
		var1 = float64(calib_press9) * pressure * pressure / 2147483648.0
		var2 = pressure * float64(calib_press8) / 32768.0
		pressure = pressure + (var1+var2+float64(calib_press7))/16.0
		if pressure < 30000.0 {
			pressure = 30000.0
		}
		if pressure > 110000.0 {
			pressure = 110000.0
		}
	}

	return pressure
}

func (b *BME280) Update() map[string]float64 {
	// Temperature
	b.dev.ReadReg(temp_msb, bufTemp)
	rawTempValue := int64(bufTemp[0])<<12 | int64(bufTemp[1])<<4 | int64(bufTemp[2])>>4
	b.data[conf.TemperatureMetricsName] = b.calibrateTemp(rawTempValue)

	// Humidity
	b.dev.ReadReg(hum_msb, bufHumid)
	rawHumidValue := int64(bufHumid[0])<<8 | int64(bufHumid[1])
	b.data[conf.HumidityMetricsName] = b.calibrateHumid(rawHumidValue)

	// Pressure
	b.dev.ReadReg(press_msb, bufPress)
	rawPressValue := int64(bufPress[0])<<12 | int64(bufPress[1])<<4 | int64(bufPress[2])>>4
	b.data[conf.PressureMetricsName] = b.calibratePress(rawPressValue) / 100.0 // Convert [Pa] to [hPa]

	return b.data
}

func (b *BME280) GetConsoleHeader() string {
	return " Temperature[°C] | Humidity[%] | Pressure[hPa] "
}

func (b *BME280) GetConsoleData() string {
	msg := fmt.Sprintf(" %15.2f | %11.2f | %13.2f ",
		b.data[conf.TemperatureMetricsName], b.data[conf.HumidityMetricsName], b.data[conf.PressureMetricsName])
	return msg
}
