//go:build windows
// +build windows

package open_hardware_monitor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateHardwareQuery_EmptyConfig_ShouldSelectStar(t *testing.T) {
	p := Config{}

	expected := "SELECT * FROM HARDWARE"

	actual := p.CreateHardwareQuery()

	assert.Equal(t, expected, actual)
}

func Test_CreateHardwareQuery_WithOneFilterValue_ShouldSelectWhereFieldEqualsValue(t *testing.T) {
	p := Config{
		Hardware: []string{"/amdcpu/0"},
	}

	expected := "SELECT * FROM HARDWARE WHERE (Identifier='/amdcpu/0')"

	actual := p.CreateHardwareQuery()

	assert.Equal(t, expected, actual)
}

func Test_CreateHardwareQuery_WithMultipleFilterValues_ShouldSelectWithOrs(t *testing.T) {
	p := Config{
		Hardware: []string{"/amdcpu/0", "/hdd/1"},
	}

	expected := "SELECT * FROM HARDWARE WHERE (Identifier='/amdcpu/0' OR Identifier='/hdd/1')"

	actual := p.CreateHardwareQuery()

	assert.Equal(t, expected, actual)
}

func Test_CreateHardwareQuery_WithMultipleFilters_ShouldSelectWithOrsAndAnds(t *testing.T) {
	p := Config{
		Hardware: []string{"/amdcpu/0", "/hdd/1"},
		HardwareType: []string{"Mainboard", "CPU", "GpuNvidia"},
	}

	expected := "SELECT * FROM HARDWARE WHERE (Identifier='/amdcpu/0' OR Identifier='/hdd/1') AND (HardwareType='Mainboard' OR HardwareType='CPU' OR HardwareType='GpuNvidia')"

	actual := p.CreateHardwareQuery()

	assert.Equal(t, expected, actual)
}

func Test_CreateSensorsQuery_WithOneFilterValue_ShouldSelectWhereFieldEqualsValue(t *testing.T) {
	p := Config{
		Sensor: []string{"/hdd/0/temperature/0"},
	}

	expected := "SELECT * FROM SENSOR WHERE (Identifier='/hdd/0/temperature/0')"

	actual := p.CreateSensorsQuery()

	assert.Equal(t, expected, actual)
}

func Test_CreateSensorsQuery_WithMultipleFilterValues_ShouldSelectWithOrs(t *testing.T) {
	p := Config{
		Sensor: []string{"/hdd/0/temperature/0", "/hdd/1/temperature/0"},
	}

	expected := "SELECT * FROM SENSOR WHERE (Identifier='/hdd/0/temperature/0' OR Identifier='/hdd/1/temperature/0')"

	actual := p.CreateSensorsQuery()

	assert.Equal(t, expected, actual)
}

func Test_CreateSensorsQuery_WithMultipleFilters_ShouldSelectWithOrsAndAnds(t *testing.T) {
	p := Config{
		Sensor: []string{"/hdd/0/temperature/0", "/hdd/1/temperature/0"},
		HardwareType: []string{"Mainboard", "CPU", "GpuNvidia"},
	}

	expected := "SELECT * FROM SENSOR WHERE (Identifier='/hdd/0/temperature/0' OR Identifier='/hdd/1/temperature/0') AND (HardwareType='Mainboard' OR HardwareType='CPU' OR HardwareType='GpuNvidia')"

	actual := p.CreateSensorsQuery()

	assert.Equal(t, expected, actual)
}

func Test_BuildTelegrafData(t *testing.T) {
	// Copied straight from WMI
	sensor := Sensor{
		Identifier: "/nvidiagpu/0/load/4",
		Index: 4,
		InstanceId: "3873",
		Name: "GPU Memory",
		Parent: "/nvidiagpu/0",
		SensorType: "Load",
		Value: 33.27344,
	}

	// Copied straight from WMI
	hardware := Hardware{
		HardwareType: "GpuNvidia",
		Identifier: "/nvidiagpu/0",
		InstanceId: "3877",
		Name: "NVIDIA NVIDIA GeForce RTX 2060",
	}

	expectedTags := map[string]string{
		"Sensor_Identifier": "/nvidiagpu/0/load/4",
		"Sensor_Index": "4",
		"Sensor_InstanceId": "3873",
		"Sensor_Name": "GPU Memory",
		"Sensor_Parent": "/nvidiagpu/0",

		"Hardware_HardwareType": "GpuNvidia",
		"Hardware_Identifier": "/nvidiagpu/0",
		"Hardware_InstanceId": "3877",
		"Hardware_Name": "NVIDIA NVIDIA GeForce RTX 2060",
	}
	expectedFields := map[string]interface{}{
		"Load": 33.27344,
	}

	fields, tags := BuildTelegrafData(sensor, hardware)

	fmt.Println(expectedFields)
	fmt.Println(fields)

	fmt.Println()

	fmt.Println(expectedTags)
	fmt.Println(tags)

	assert.Equal(t, fmt.Sprint(expectedFields), fmt.Sprint(fields))
}
