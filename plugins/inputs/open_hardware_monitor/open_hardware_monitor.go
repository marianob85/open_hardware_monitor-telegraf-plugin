//go:build windows
// +build windows

package open_hardware_monitor

import (
	"fmt"
	"log"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type OpenHardwareMonitorConfig struct {
	SensorsType []string
	Hardware    []string
}

type OpenHardwareMonitorSensor struct {
	Identifier string
	Index      int32
	InstanceId string
	Max        float32
	Min        float32
	Name       string
	Parent     string
	ProcessId  string
	SensorType string
	Value      float32
}

type OpenHardwareMonitorHardware struct {
	HardwareType string
	Identifier   string
	InstanceId   string
	Name         string
	ProcessId    string
}

type OpenHardwareMonitorData struct {
	Hardware []OpenHardwareMonitorHardware
	Sensors []OpenHardwareMonitorSensor
}

func (p *OpenHardwareMonitorConfig) Description() string {
	return "Get sensors data from Open Hardware Monitor via WMI"
}

const sampleConfig = `
	# Which types of sensors should metrics be collected from
	# If not given, all sensor types are included
	SensorsType = ["Temperature", "Fan", "Voltage"] # optional
	
	# Which hardware should be metrics be collected from
	# If not given, all hardware is included
	Hardware = ["/intelcpu/0"]  # optional
`

func (p *OpenHardwareMonitorConfig) SampleConfig() string {
	return sampleConfig
}

func (p *OpenHardwareMonitorConfig) CreateHardwareQuery() (string, error) {
	query := "SELECT * FROM HARDWARE"
	if len(p.Hardware) != 0 {
		query += " WHERE "
		var hardware []string
		for _, h := range p.Hardware {
			hardware = append(hardware, fmt.Sprint("HardwareType='", h, "'"))
		}
		query += strings.Join(hardware, " OR ")
	}
	return query, nil
}

func (p *OpenHardwareMonitorConfig) QueryHardware() ([]OpenHardwareMonitorHardware, error) {
	hardwareQuery, err := p.CreateHardwareQuery()
	if err != nil {
		log.Fatal(err)
	}

	var hardware []OpenHardwareMonitorHardware
	err = wmi.QueryNamespace(hardwareQuery, &hardware, "root/OpenHardwareMonitor")

	// TODO remove debug logging
	fmt.Println("=== Hardware ===")
	for i := range hardware {
		fmt.Println(hardware[i])
	}

	return hardware, err
}

func (p *OpenHardwareMonitorConfig) CreateSensorsQuery() (string, error) {
	query := "SELECT * FROM SENSOR"
	if len(p.SensorsType) != 0 {
		query += " WHERE "
		var sensors []string
		for _, sensor := range p.SensorsType {
			sensors = append(sensors, fmt.Sprint("SensorType='", sensor, "'"))
		}
		query += strings.Join(sensors, " OR ")
	}
	return query, nil
}

func (p *OpenHardwareMonitorConfig) QuerySensors() ([]OpenHardwareMonitorSensor, error) {
	sensorsQuery, err := p.CreateSensorsQuery()
	if err != nil {
		log.Fatal(err)
	}

	var sensors []OpenHardwareMonitorSensor
	err = wmi.QueryNamespace(sensorsQuery, &sensors, "root/OpenHardwareMonitor")

	// TODO remove debug logging
	fmt.Println("=== Sensors ===")
	for i := range sensors {
		fmt.Println(sensors[i])
	}

	return sensors, err
}

func contains(s []string, e string) bool {
	if len(s) == 0 {
		return true
	}
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (p *OpenHardwareMonitorConfig) Gather(acc telegraf.Accumulator) error {
	p.QueryHardware()

	var dst []OpenHardwareMonitorSensor
	dst, err := p.QuerySensors()

	if err != nil {
		acc.AddError(err)
	}

	for _, sensorData := range dst {
		if contains(p.Hardware, sensorData.Parent) {
			tags := map[string]string{
				"name":   sensorData.Name,
				"parent": sensorData.Parent,
			}
			fields := map[string]interface{}{sensorData.SensorType: sensorData.Value}
			acc.AddFields("ohm", fields, tags)
		}
	}


	return nil
}

func init() {
	inputs.Add("open_hardware_monitor", func() telegraf.Input {
		return &OpenHardwareMonitorConfig{}
	})
}
