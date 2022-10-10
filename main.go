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
)

const namespace = "stecagrid"

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client    = &http.Client{Transport: tr}
	frequency = 5

	listenAddress = flag.String("web.listen-address", ":9141",
		"Address to listen on for telemetry")
	metricsPath = flag.String("web.telemetry-path", "/metrics",
		"Path under which to expose metrics")

	// StecaGrid Metrics
	acPower = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "ac_power"),
		"How many messages have been received (per channel).",
		[]string{"channel"}, nil,
	)
	acVoltage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "ac_voltage"),
		"How many messages have been filtered (per channel).",
		[]string{"channel"}, nil,
	)
	// more metrics...
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

func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}

func main() {
	// http.Handle("/metrics", promhttp.Handler())
	// log.Fatal(http.ListenAndServe(":9101", nil))

	for {
		if xmlBytes, err := getXML("http://192.168.50.144/measurements.xml"); err != nil {
			log.Printf("Failed to get XML: %v", err)
		} else {
			var results stecaGrid
			err = xml.Unmarshal(xmlBytes, &results)
			if err != nil {
				log.Fatalf("xml.Unmarshal failed with '%s'\n", err)
			}
			fmt.Println(results.Device.Measurements.Measurement[0].Value)

			// TODO ingest prometheus metrics

		}
		time.Sleep(time.Second * time.Duration(frequency))
	}
}
