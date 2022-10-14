package main

import (
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	stecaIP   = flag.String("steca-ip", "192.168.50.144", "StecaGrid IP address")
	stecaPath = flag.String("steca-path", "/measurements.xml", "StecaGrid path")
	namespace = "stecagrid"
	tr        = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClientTimeout = time.Duration(3 * time.Second)
	httpClient        = &http.Client{
		Transport: tr,
		Timeout:   httpClientTimeout,
	}
	frequency = flag.Int("frequency", 5, "Polling frequency in seconds")

	// StecaGrid Metrics
	acPower = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ac_power",
			Help:      "AC Power (W)",
		})
	acCurrent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ac_current",
			Help:      "AC Current (A)",
		})
	temperature = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "temp",
			Help:      "Temperature (°C)",
		})
	acVoltage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ac_voltage",
			Help:      "AC Voltage (V)",
		})
	acFrequency = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ac_frequency",
			Help:      "AC Frequency (Hz)",
		})
	gridPower = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "grid_power",
			Help:      "Grid Power (W)",
		})
	dcVoltage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "dc_voltage",
			Help:      "DC Voltage (V)",
		})
	dcCurrent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "dc_current",
			Help:      "DC Current (A)",
		})
	derating = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "derating",
			Help:      "Derating (%)",
		})
)

type stecaGrid struct {
	XMLName xml.Name `xml:"root"`
	Text    string   `xml:",chardata"`
	Device  struct {
		Text         string `xml:",chardata"`
		Name         string `xml:"Name,attr"`
		NominalPower string `xml:"NominalPower,attr"`
		Type         string `xml:"Type,attr"`
		Serial       string `xml:"Serial,attr"`
		BusAddress   string `xml:"BusAddress,attr"`
		NetBiosName  string `xml:"NetBiosName,attr"`
		IPAddress    string `xml:"IpAddress,attr"`
		DateTime     string `xml:"DateTime,attr"`
		Measurements struct {
			Text        string `xml:",chardata"`
			Measurement []struct {
				Text  string  `xml:",chardata"`
				Value float64 `xml:"Value,attr"`
				Unit  string  `xml:"Unit,attr"`
				Type  string  `xml:"Type,attr"`
			} `xml:"Measurement"`
		} `xml:"Measurements"`
	} `xml:"Device"`
}

func setupPrometheus() {
	prometheus.MustRegister(acPower)
	prometheus.MustRegister(acVoltage)
	prometheus.MustRegister(dcVoltage)
	prometheus.MustRegister(acCurrent)
	prometheus.MustRegister(dcCurrent)
	prometheus.MustRegister(acFrequency)
	prometheus.MustRegister(temperature)
	prometheus.MustRegister(gridPower)
	prometheus.MustRegister(derating)
	go func() {
		for {
			http.Handle("/metrics", promhttp.Handler())
			log.Fatal(http.ListenAndServe(":9101", nil))
		}
	}()
}

func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read body error: %v", err)
	}

	return data, nil
}

func main() {
	setupPrometheus()
	flag.Parse()

	stecaURL := "http://" + *stecaIP + *stecaPath

	for {
		if xmlBytes, err := getXML(stecaURL); err != nil {
			log.Printf("Failed to fetch the metrics: %v", err)
		} else {
			var results stecaGrid
			if err := xml.Unmarshal(xmlBytes, &results); err != nil {
				log.Fatal(err)
			}

			for k := range results.Device.Measurements.Measurement {
				t := results.Device.Measurements.Measurement[k].Type
				switch t {
				case "AC_Voltage":
					acVoltage.Set(results.Device.Measurements.Measurement[k].Value)
				case "AC_Current":
					acCurrent.Set(results.Device.Measurements.Measurement[k].Value)
				case "AC_Power":
					acPower.Set(results.Device.Measurements.Measurement[k].Value)
				case "AC_Frequency":
					acFrequency.Set(results.Device.Measurements.Measurement[k].Value)
				case "DC_Voltage":
					dcVoltage.Set(results.Device.Measurements.Measurement[k].Value)
				case "DC_Current":
					dcCurrent.Set(results.Device.Measurements.Measurement[k].Value)
				case "Temp":
					temperature.Set(results.Device.Measurements.Measurement[k].Value)
				case "GridPower":
					gridPower.Set(results.Device.Measurements.Measurement[k].Value)
				case "Derating":
					derating.Set(results.Device.Measurements.Measurement[k].Value)
				}
			}

		}
		time.Sleep(time.Second * time.Duration(*frequency))
	}
}
