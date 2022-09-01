//go:build windows
// +build windows

package open_hardware_monitor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Config struct {
	HardwareType []string
	Hardware     []string
	SensorType   []string
	Sensor       []string
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
	HardwareType = ["CPU", "GpuNvidia"] # optional

	# Which hardware identifiers should metrics be collected from
	# If not given, all hardware is included
	Hardware = ["/intelcpu/0", "/nvidiagpu/0"]  # optional

	# Which types of sensors should metrics be collected from
	# If not given, all sensor types are included
	SensorType = ["Temperature", "Fan", "Voltage"] # optional

	# Which hardware identifiers should metrics be collected from
	# If not given, all hardware is included
	Sensor = ["/intelcpu/0", "/nvidiagpu/0"]  # optional
	
`

func (p *Config) SampleConfig() string {
	return sampleConfig
}

func BuildOrConditions(values []string, field string) string {
	conditions := []string{}
	for _, value := range values {
		conditions = append(conditions, fmt.Sprintf("%s='%s'", field, value))
	}

	if len(conditions) != 0 {
		return "(" + strings.Join(conditions, " OR ") + ")"
	} else {
		return ""
	}
}

func BuildAndConditions(conditions []string) string {
	andConditions := []string{}
	for _, condition := range conditions {
		if len(condition) != 0 {
			andConditions = append(andConditions, condition)
		}
	}
	return strings.Join(andConditions, " AND ")
}

func (p *Config) CreateHardwareQuery() string {
	query := "SELECT * FROM HARDWARE"

	hardwareConditions := BuildOrConditions(p.Hardware, "Identifier")
	hardwareTypeConditions := BuildOrConditions(p.HardwareType, "HardwareType")

	allConditions := BuildAndConditions([]string{
		hardwareConditions,
		hardwareTypeConditions,
	})

	if len(allConditions) != 0 {
		query += " WHERE " + allConditions
	}

	return query
}

func (p *Config) QueryHardware() ([]Hardware, error) {
	hardwareQuery := p.CreateHardwareQuery()

	var hardware []Hardware
	err := wmi.QueryNamespace(hardwareQuery, &hardware, "root/OpenHardwareMonitor")

	return hardware, err
}

func (p *Config) CreateSensorsQuery() string {
	query := "SELECT * FROM SENSOR"

	sensorConditions := BuildOrConditions(p.Sensor, "Identifier")
	sensorTypeConditions := BuildOrConditions(p.SensorType, "SensorType")
	hardwareConditions := BuildOrConditions(p.Hardware, "Parent")
	hardwareTypeConditions := BuildOrConditions(p.HardwareType, "HardwareType")

	allConditions := BuildAndConditions([]string{
		sensorConditions,
		sensorTypeConditions,
		hardwareConditions,
		hardwareTypeConditions,
	})

	if len(allConditions) != 0 {
		query += " WHERE " + allConditions
	}

	return query
}

func (p *Config) QuerySensors() ([]Sensor, error) {
	sensorsQuery := p.CreateSensorsQuery()

	var sensors []Sensor
	err := wmi.QueryNamespace(sensorsQuery, &sensors, "root/OpenHardwareMonitor")

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
		} else {
			acc.AddError(
				fmt.Errorf(
					"unable to find hardware associated with sensor '%s', metrics will not be collected",
					s.Identifier,
				),
			)
		}
	}

	return nil
}

func init() {
	inputs.Add("open_hardware_monitor", func() telegraf.Input {
		return &Config{}
	})
}
