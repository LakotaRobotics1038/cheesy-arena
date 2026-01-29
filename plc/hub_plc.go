// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for interfacing with hub-specific field PLCs.

package plc

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
)

// HubPlc represents a PLC dedicated to controlling a single hub (red or blue).
type HubPlc interface {
	SetAddress(address string)
	IsEnabled() bool
	IsHealthy() bool
	IoChangeNotifier() *websocket.Notifier
	Run()
	ResetMatch()
	GetHubCount() int
	SetHubCount(count int)
	SetHubLight(active bool)
	SetHubActive(active bool)
	GetInputNames() []string
	GetRegisterNames() []string
	GetCoilNames() []string
}

// ModbusHubPlc implements the HubPlc interface for a single hub.
type ModbusHubPlc struct {
	address           string
	handler           *modbus.TCPClientHandler
	client            modbus.Client
	isHealthy         bool
	ioChangeNotifier  *websocket.Notifier
	registers         [hubRegisterCount]uint16
	coils             [hubCoilCount]bool
	oldRegisters      [hubRegisterCount]uint16
	oldCoils          [hubCoilCount]bool
	hubActiveRegister bool
	allianceName      string // "red" or "blue" for logging purposes
	cycleCounter      int
}

const (
	hubPlcLoopPeriodMs    = 100
	hubPlcRetryIntevalSec = 3
	hubCycleCounterMax    = 100
)

// Hub-specific registers
//
//go:generate stringer -type=hubRegister
type hubRegister int

const (
	hubIoConnection hubRegister = iota
	hubLightRed
	hubLightBlue
	hubLightGreen
	hubActive
	hubCount
	hubRegisterCount
)

// Hub-specific coils
//
//go:generate stringer -type=hubCoil
type hubCoil int

const (
	hubHeartbeat hubCoil = iota
	hubCoilCount
)

// NewHubPlc creates a new ModbusHubPlc instance for the given alliance.
func NewHubPlc(allianceName string) *ModbusHubPlc {
	plc := new(ModbusHubPlc)
	plc.allianceName = allianceName
	plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", nil)
	return plc
}

func (plc *ModbusHubPlc) SetAddress(address string) {
	if address != plc.address {
		plc.address = address
		plc.resetConnection()
	}
}

func (plc *ModbusHubPlc) IsEnabled() bool {
	return plc.address != ""
}

func (plc *ModbusHubPlc) IsHealthy() bool {
	return plc.isHealthy
}

func (plc *ModbusHubPlc) IoChangeNotifier() *websocket.Notifier {
	return plc.ioChangeNotifier
}

func (plc *ModbusHubPlc) Run() {
	for {
		if plc.IsEnabled() {
			if plc.client == nil {
				err := plc.connect()
				if err != nil {
					log.Printf("%s hub PLC: %v", plc.allianceName, err)
					plc.isHealthy = false
					time.Sleep(hubPlcRetryIntevalSec * time.Second)
					continue
				}
			}

			plc.update()
		}

		time.Sleep(hubPlcLoopPeriodMs * time.Millisecond)
	}
}

func (plc *ModbusHubPlc) GetInputNames() []string {
	return []string{} // Hub PLCs don't have inputs
}

func (plc *ModbusHubPlc) GetRegisterNames() []string {
	registerNames := make([]string, uint16(hubRegisterCount))
	for i := range plc.registers {
		registerNames[i] = hubRegister(i).String()
	}
	return registerNames
}

func (plc *ModbusHubPlc) GetCoilNames() []string {
	coilNames := make([]string, uint16(hubCoilCount))
	for i := range plc.coils {
		coilNames[i] = hubCoil(i).String()
	}
	return coilNames
}

// ResetMatch resets the hub count, active state, and lights for a new match.
func (plc *ModbusHubPlc) ResetMatch() {
	plc.SetHubCount(0)
	plc.SetHubActive(false)
	plc.SetHubLight(false)
}

// GetHubCount returns the current hub count for this alliance.
func (plc *ModbusHubPlc) GetHubCount() int {
	return int(plc.registers[hubCount])
}

// SetHubCount sets the hub count register (not typically called - PLC updates this).
func (plc *ModbusHubPlc) SetHubCount(count int) {
	plc.registers[hubCount] = uint16(count)
}

// SetHubLight sets the state of the hub LED light.
func (plc *ModbusHubPlc) SetHubLight(active bool) {
	if active {
		switch plc.allianceName {
		case "red":
			plc.registers[hubLightRed] = 255
			plc.registers[hubLightGreen] = 0
			plc.registers[hubLightBlue] = 0
		case "blue":
			plc.registers[hubLightRed] = 0
			plc.registers[hubLightGreen] = 0
			plc.registers[hubLightBlue] = 255
		default:
			plc.registers[hubLightRed] = 255
			plc.registers[hubLightGreen] = 255
			plc.registers[hubLightBlue] = 255
		}
	} else {
		// Turn off all lights
		plc.registers[hubLightRed] = 0
		plc.registers[hubLightGreen] = 0
		plc.registers[hubLightBlue] = 0
	}
}

// SetHubActive sets the active state for the hub and controls the LED light.
func (plc *ModbusHubPlc) SetHubActive(active bool) {
	plc.hubActiveRegister = active
	if active {
		plc.registers[hubActive] = 1
	} else {
		plc.registers[hubActive] = 0
	}
}

func (plc *ModbusHubPlc) connect() error {
	if plc.address == "" {
		return fmt.Errorf("%s hub PLC address is not set", plc.allianceName)
	}

	address := plc.address
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%s:%d", address, modbusPort)
	}
	plc.handler = modbus.NewTCPClientHandler(address)
	plc.handler.Timeout = 1 * time.Second

	err := plc.handler.Connect()
	if err != nil {
		plc.resetConnection()
		return err
	}
	plc.client = modbus.NewClient(plc.handler)
	log.Printf("%s hub PLC connected to %s", plc.allianceName, plc.address)
	return nil
}

func (plc *ModbusHubPlc) resetConnection() {
	if plc.handler != nil {
		plc.handler.Close()
	}
	plc.handler = nil
	plc.client = nil
}

// Performs a single iteration of reading inputs from and writing outputs to the hub PLC.
func (plc *ModbusHubPlc) update() {
	// Update heartbeat
	plc.cycleCounter = (plc.cycleCounter + 1) % hubCycleCounterMax
	plc.coils[hubHeartbeat] = plc.cycleCounter < hubCycleCounterMax/2

	plc.isHealthy = plc.readRegisters() && plc.writeRegisters() && plc.writeCoils()
	if !plc.isHealthy {
		plc.resetConnection()
	}
	plc.checkForChanges()
}

func (plc *ModbusHubPlc) readRegisters() bool {
	if plc.client == nil {
		return false
	}

	registers, err := plc.client.ReadHoldingRegisters(0, uint16(hubRegisterCount))
	if err != nil {
		log.Printf("%s hub PLC error reading registers: %v", plc.allianceName, err)
		return false
	}

	plc.registers[hubIoConnection] = uint16(registers[2*hubIoConnection])<<8 + uint16(registers[2*hubIoConnection+1])
	plc.registers[hubCount] = uint16(registers[2*hubCount])<<8 + uint16(registers[2*hubCount+1])

	return true
}

func (plc *ModbusHubPlc) writeRegisters() bool {
	if plc.client == nil {
		return false
	}

	registerBytes := make([]byte, 2*uint16(hubRegisterCount))
	for i, register := range plc.registers {
		registerBytes[2*i] = byte(register >> 8)
		registerBytes[2*i+1] = byte(register)
	}

	_, err := plc.client.WriteMultipleRegisters(0, uint16(hubRegisterCount), registerBytes)
	if err != nil {
		log.Printf("%s hub PLC error writing registers: %v", plc.allianceName, err)
		return false
	}
	return true
}

func (plc *ModbusHubPlc) writeCoils() bool {
	if plc.client == nil {
		return false
	}

	coilBytes := make([]byte, (hubCoilCount+7)/8)
	for i, coil := range plc.coils {
		if coil {
			coilBytes[i/8] |= 1 << uint(i%8)
		}
	}

	_, err := plc.client.WriteMultipleCoils(0, uint16(hubCoilCount), coilBytes)
	if err != nil {
		log.Printf("%s hub PLC error writing coils: %v", plc.allianceName, err)
		return false
	}
	return true
}

func (plc *ModbusHubPlc) checkForChanges() {
	// Check if any registers have changed.
	for i, register := range plc.registers {
		if register != plc.oldRegisters[i] {
			plc.ioChangeNotifier.Notify()
			plc.oldRegisters = plc.registers
			break
		}
	}

	// Check if any coils have changed.
	for i, coil := range plc.coils {
		if coil != plc.oldCoils[i] {
			plc.ioChangeNotifier.Notify()
			plc.oldCoils = plc.coils
			break
		}
	}
}
