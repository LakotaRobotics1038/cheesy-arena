// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package plc

import (
	"testing"

	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
)

func TestHubPlcInitialization(t *testing.T) {
	plc := NewHubPlc("red")
	assert.Equal(t, "red", plc.allianceName)
	assert.NotNil(t, plc.ioChangeNotifier)
	assert.Equal(t, "", plc.address)
	assert.Equal(t, false, plc.IsEnabled())
	assert.Equal(t, false, plc.IsHealthy())
}

func TestHubPlcSetAddress(t *testing.T) {
	plc := NewHubPlc("blue")
	assert.Equal(t, false, plc.IsEnabled())

	plc.SetAddress("10.0.100.50")
	assert.Equal(t, true, plc.IsEnabled())
	assert.Equal(t, "10.0.100.50", plc.address)
}

func TestHubPlcGetHubCount(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusHubPlc
	plc.allianceName = "red"
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)

	client.registers[int(hubCount)] = 0
	plc.update()
	assert.Equal(t, 0, plc.GetHubCount())

	client.registers[int(hubCount)] = 15
	plc.update()
	assert.Equal(t, 15, plc.GetHubCount())

	client.registers[int(hubCount)] = 42
	plc.update()
	assert.Equal(t, 42, plc.GetHubCount())
}

func TestHubPlcSetHubCount(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusHubPlc
	plc.allianceName = "blue"
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)

	// SetHubCount writes to the PLC, but update() reads it back
	// So we need to set it on both sides for testing
	plc.SetHubCount(10)
	client.registers[int(hubCount)] = 10
	plc.update()
	assert.Equal(t, 10, plc.GetHubCount())

	plc.SetHubCount(25)
	client.registers[int(hubCount)] = 25
	plc.update()
	assert.Equal(t, 25, plc.GetHubCount())
}

func TestHubPlcSetHubActive(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusHubPlc
	plc.allianceName = "red"
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)

	plc.SetHubActive(false)
	plc.update()
	assert.Equal(t, uint16(0), client.registers[int(hubActive)])

	plc.SetHubActive(true)
	plc.update()
	assert.Equal(t, uint16(1), client.registers[int(hubActive)])

	plc.SetHubActive(false)
	plc.update()
	assert.Equal(t, uint16(0), client.registers[int(hubActive)])
}

func TestHubPlcSetHubLightRed(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusHubPlc
	plc.allianceName = "red"
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)

	plc.SetHubLight(false)
	plc.update()
	assert.Equal(t, uint16(0), client.registers[int(hubLightRed)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightGreen)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightBlue)])

	plc.SetHubLight(true)
	plc.update()
	assert.Equal(t, uint16(255), client.registers[int(hubLightRed)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightGreen)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightBlue)])

	plc.SetHubLight(false)
	plc.update()
	assert.Equal(t, uint16(0), client.registers[int(hubLightRed)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightGreen)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightBlue)])
}

func TestHubPlcSetHubLightBlue(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusHubPlc
	plc.allianceName = "blue"
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)

	plc.SetHubLight(false)
	plc.update()
	assert.Equal(t, uint16(0), client.registers[int(hubLightRed)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightGreen)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightBlue)])

	plc.SetHubLight(true)
	plc.update()
	assert.Equal(t, uint16(0), client.registers[int(hubLightRed)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightGreen)])
	assert.Equal(t, uint16(255), client.registers[int(hubLightBlue)])

	plc.SetHubLight(false)
	plc.update()
	assert.Equal(t, uint16(0), client.registers[int(hubLightRed)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightGreen)])
	assert.Equal(t, uint16(0), client.registers[int(hubLightBlue)])
}

func TestHubPlcHeartbeat(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusHubPlc
	plc.allianceName = "red"
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)

	assert.Equal(t, false, client.coils[int(hubHeartbeat)])
	plc.update()
	assert.Equal(t, true, client.coils[int(hubHeartbeat)])

	// Should stay true for cycles 2-49 (48 more updates)
	for i := 0; i < hubCycleCounterMax/2-2; i++ {
		plc.update()
	}
	assert.Equal(t, true, client.coils[int(hubHeartbeat)])

	// At cycle 50, should become false
	plc.update()
	assert.Equal(t, false, client.coils[int(hubHeartbeat)])

	// Should stay false for cycles 51-99 (48 more updates)
	for i := 0; i < hubCycleCounterMax/2-2; i++ {
		plc.update()
	}
	assert.Equal(t, false, client.coils[int(hubHeartbeat)])

	// At cycle 99, still false
	plc.update()
	assert.Equal(t, false, client.coils[int(hubHeartbeat)])

	// At cycle 0 (wraps around), should become true again
	plc.update()
	assert.Equal(t, true, client.coils[int(hubHeartbeat)])
}

func TestHubPlcGetNames(t *testing.T) {
	plc := NewHubPlc("red")

	inputNames := plc.GetInputNames()
	assert.Equal(t, 0, len(inputNames))

	registerNames := plc.GetRegisterNames()
	assert.Equal(t, int(hubRegisterCount), len(registerNames))
	assert.Equal(t, "hubIoConnection", registerNames[0])
	assert.Equal(t, "hubLightRed", registerNames[1])
	assert.Equal(t, "hubCount", registerNames[5])

	coilNames := plc.GetCoilNames()
	assert.Equal(t, int(hubCoilCount), len(coilNames))
	assert.Equal(t, "hubHeartbeat", coilNames[0])
}

func TestHubPlcIsHealthy(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusHubPlc
	plc.allianceName = "blue"
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)

	plc.update()
	assert.Equal(t, true, plc.IsHealthy())

	client.returnError = true
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())

	client.returnError = false
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())
}
