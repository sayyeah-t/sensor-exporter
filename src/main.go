package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"sensor-exporter/config"
	"sensor-exporter/sensor"
)

var (
	// args
	outputStdout int
	confPath     string
	conf         config.Default
	sensors      []sensor.Sensor
	data         map[string]float64
	headerData   string
	headerCount  int = 0
)

func UpdateData() {
	var msg []string
	for _, s := range sensors {
		sensorData := s.Update()
		for i, d := range sensorData {
			data[i] = d
			setExportValue(i, s.GetSensorName())
		}
		msg = append(msg, s.GetConsoleData())
	}
	if headerCount == 0 {
		if outputStdout != 0 {
			log.Println(headerData)
		}
	}
	headerCount = (headerCount + 1) % 15
	if outputStdout != 0 {
		log.Println("|" + strings.Join(msg, "|") + "|")
	}
}

func main() {
	// parse arguments
	flag.IntVar(&outputStdout, "stdout", 0, "1: output sensor data to stdout, 0: do not it")
	flag.StringVar(&confPath, "config", "/etc/sensor-exporter/sensor-exporter.conf", "config file")
	flag.Parse()

	// make channel for stop application
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	// init configuration data
	err := config.Init(confPath)
	if err != nil {
		log.Printf("Config init error: %v\n", err)
		os.Exit(1)
	}
	config.DumpConfig()
	conf = config.GetConfig().Default

	// init metrics
	data = make(map[string]float64, len(conf.ExportMetrics))
	for _, s := range conf.ExportMetrics {
		data[s] = 0.0
	}

	// init sensors and stdout header string
	sensors = sensor.Init(conf.EnabledSensors)
	tmpHeaderData := make([]string, len(sensors))
	for i, s := range sensors {
		initerr := s.Init()
		if initerr != nil {
			log.Printf("sensor init error: %v\n", initerr)
			os.Exit(1)
		}
		tmpHeaderData[i] = s.GetConsoleHeader()
	}
	headerData = "|" + strings.Join(tmpHeaderData, "|") + "|"

	// init prometheus exporter
	initExporter()

	// define a function for stop application
	defer func() {
		for _, s := range sensors {
			s.Close()
		}
		stopExporter()
		close(sig)
	}()

	// init ticker for update metrics every 1 sec
	ticker := time.NewTicker(1 * time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		flag := true
		for {
			select {
			case s := <-sig:
				log.Printf("Received signal: %v\n", s)
				flag = false
			case <-ticker.C:
				UpdateData()
			}
			if !flag {
				break
			}
		}
		wg.Done()
	}()

	// run prometheus exporter
	go runExporter()

	// wait to stop update metrics
	wg.Wait()

	log.Println("Stop application")
}
