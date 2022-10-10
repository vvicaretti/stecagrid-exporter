package main

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	stecaIP   = "192.168.50.144"
	stecaPath = "/measurements.xml"
	namespace = "stecagrid"
	tr        = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClientTimeout = time.Duration(3 * time.Second)
	httpClient        = &http.Client{
		Transport: tr,
		Timeout:   httpClientTimeout,
	}
	frequency = 10
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

/* Example of output from http://<stecagrid-ip>/measurements.xml

<root>
  <Device Name="StecaGrid 2000" NominalPower="2000" Type="Inverter" Serial="<xxxxxx>" BusAddress="1" NetBiosName="xxxxx" IpAddress="x.x.x.x" DateTime="2021-08-03T21:39:21">
  <Measurements>
    <Measurement Value="232.8" Unit="V" Type="AC_Voltage"/>
    <Measurement Unit="A" Type="AC_Current"/>
	<Measurement Unit="W" Type="AC_Power"/>
	<Measurement Value="50.131" Unit="Hz" Type="AC_Frequency"/>
	<Measurement Value="1.1" Unit="V" Type="DC_Voltage"/>
	<Measurement Unit="A" Type="DC_Current"/>
	<Measurement Unit="°C" Type="Temp"/>
	<Measurement Unit="W" Type="GridPower"/>
	<Measurement Value="100.0" Unit="%" Type="Derating"/>
  </Measurements>
  </Device>
</root>

*/

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
				Text  string `xml:",chardata"`
				Value string `xml:"Value,attr"`
				Unit  string `xml:"Unit,attr"`
				Type  string `xml:"Type,attr"`
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
		return []byte{}, fmt.Errorf("read body: %v", err)
	}

	return data, nil
}

func main() {
	setupPrometheus()

	stecaURL := "http://" + stecaIP + stecaPath

	for {
		if xmlBytes, err := getXML(stecaURL); err != nil {
			log.Printf("Failed to get XML: %v", err)
		} else {
			var results stecaGrid
			if err := xml.Unmarshal(xmlBytes, &results); err != nil {
				log.Fatal(err)
			}
			// AC Power
			intPower, err := strconv.ParseFloat(results.Device.Measurements.Measurement[2].Value, 64)
			if err == nil {
				acPower.Set(intPower)
			}
			// AC Voltage
			intACVoltage, err := strconv.ParseFloat(results.Device.Measurements.Measurement[0].Value, 64)
			if err == nil {
				acVoltage.Set(intACVoltage)
			}
			// AC Frequency
			intFrequency, err := strconv.ParseFloat(results.Device.Measurements.Measurement[3].Value, 64)
			if err == nil {
				acFrequency.Set(intFrequency)
			}
			// AC Current
			intACCurrent, err := strconv.ParseFloat(results.Device.Measurements.Measurement[1].Value, 64)
			if err == nil {
				acFrequency.Set(intACCurrent)
			}
			// DC Current
			intDCCurrent, err := strconv.ParseFloat(results.Device.Measurements.Measurement[5].Value, 64)
			if err == nil {
				dcCurrent.Set(intDCCurrent)
			}
			// GridPower
			intGridPower, err := strconv.ParseFloat(results.Device.Measurements.Measurement[7].Value, 64)
			if err == nil {
				acFrequency.Set(intGridPower)
			}
			// Derating
			intDerating, err := strconv.ParseFloat(results.Device.Measurements.Measurement[8].Value, 64)
			if err == nil {
				acFrequency.Set(intDerating)
			}
			// DC Voltage
			intDCVoltage, err := strconv.ParseFloat(results.Device.Measurements.Measurement[4].Value, 64)
			if err == nil {
				acFrequency.Set(intDCVoltage)
			}
			// Temp
			intTemp, err := strconv.ParseFloat(results.Device.Measurements.Measurement[6].Value, 64)
			if err == nil {
				acFrequency.Set(intTemp)
			}

		}
		time.Sleep(time.Second * time.Duration(frequency))
	}
}
