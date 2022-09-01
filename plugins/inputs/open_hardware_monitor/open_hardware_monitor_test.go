//go:build windows
// +build windows

package open_hardware_monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateQueryWithSensors(t *testing.T) {
	//var acc testutil.Accumulator
	p := Config{
		SensorsType: []string{"Temperature", "Voltage"},
	}

	query, _ := p.CreateSensorsQuery()

	assert.Equal(t, "SELECT * FROM SENSOR WHERE SensorType='Temperature' OR SensorType='Voltage'", query)
}

func TestCreateQueryEmpty(t *testing.T) {
	//var acc testutil.Accumulator
	var p Config

	query, _ := p.CreateSensorsQuery()

	assert.Equal(t, "SELECT * FROM SENSOR", query)
}
