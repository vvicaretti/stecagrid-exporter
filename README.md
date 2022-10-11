# stecagrid-exporter
StecaGrid exporter


Example of output from `http://<stecagrid-ip>/measurements.xml`

```
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
```
