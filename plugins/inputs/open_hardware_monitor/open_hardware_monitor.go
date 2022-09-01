//go:build windows
// +build windows

package open_hardware_monitor

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Config struct {
	SensorsType []string
	Hardware    []string
}

type Hardware struct {
	HardwareType string
	Identifier   string
	InstanceId   string
	Name         string
}

type Sensor struct {
	Identifier string
	Index      int
	InstanceId string
	Name       string
	Parent     string
	SensorType string
	Value      float32
}

func (p *Config) Description() string {
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

func (p *Config) SampleConfig() string {
	return sampleConfig
}

func (p *Config) CreateHardwareQuery() (string, error) {
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

func (p *Config) QueryHardware() ([]Hardware, error) {
	hardwareQuery, err := p.CreateHardwareQuery()
	if err != nil {
		log.Fatal(err)
	}

	var hardware []Hardware
	err = wmi.QueryNamespace(hardwareQuery, &hardware, "root/OpenHardwareMonitor")

	return hardware, err
}

func (p *Config) CreateSensorsQuery() (string, error) {
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

func (p *Config) QuerySensors() ([]Sensor, error) {
	sensorsQuery, err := p.CreateSensorsQuery()
	if err != nil {
		log.Fatal(err)
	}

	var sensors []Sensor
	err = wmi.QueryNamespace(sensorsQuery, &sensors, "root/OpenHardwareMonitor")

	return sensors, err
}

func BuildTelegrafData(sensor Sensor, hardware Hardware) (map[string]interface{}, map[string]string) {
	// Field is just the sensor value
	fields := map[string]interface{}{
		sensor.SensorType: sensor.Value,
	}

	tags := map[string]string{
		// Sensor info
		"Sensor_Identifier": sensor.Identifier,
		"Sensor_Index": strconv.Itoa(sensor.Index),
		"Sensor_InstanceId": sensor.InstanceId,
		"Sensor_Name": sensor.Name,
		"Sensor_Parent": sensor.Parent,

		// Hardware info
		"Hardware_HardwareType": hardware.HardwareType,
		"Hardware_Identifier": hardware.Identifier,
		"Hardware_InstanceId": hardware.InstanceId,
		"Hardware_Name": hardware.Name,
	}

	return fields, tags
}

func (p *Config) Gather(acc telegraf.Accumulator) error {
	var hardware []Hardware
	hardware, err := p.QueryHardware()

	if err != nil {
		acc.AddError(err)
	}

	hardwareByIdentifier := map[string]Hardware{}
	for _, h := range hardware {
		hardwareByIdentifier[h.Identifier] = h
	}

	var sensors []Sensor
	sensors, err = p.QuerySensors()

	if err != nil {
		acc.AddError(err)
	}

	// For each sensor
	for _, s := range sensors {
		// If sensors parent is included in hardware
		if h, exists := hardwareByIdentifier[s.Parent]; exists {
			fields, tags := BuildTelegrafData(s, h)
			acc.AddFields("openhardwaremonitor", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("open_hardware_monitor", func() telegraf.Input {
		return &Config{}
	})
}
