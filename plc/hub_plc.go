// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for interfacing with hub-specific field PLCs.

package plc

import (
	"fmt"
	"log"
	"strings"
	"sync"
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
	SetHubLightColor(red, green, blue uint16)
	SetHubAnimation(animation LedAnimation)
	SetHubChaserAnimation()
	SetHubActive(active bool)
	GetInputNames() []string
	GetRegisterNames() []string
	GetCoilNames() []string
}

// ModbusHubPlc implements the HubPlc interface for a single hub.
type ModbusHubPlc struct {
	address          string
	handler          *modbus.TCPClientHandler
	client           modbus.Client
	isHealthy        bool
	ioChangeNotifier *websocket.Notifier
	registers        [hubRegisterCount]uint16
	coils            [hubCoilCount]bool
	oldRegisters     [hubRegisterCount]uint16
	oldCoils         [hubCoilCount]bool
	allianceName     string // "red" or "blue" for logging purposes
	modbusLock       sync.Mutex
}

const (
	hubPlcLoopPeriodMs    = 100
	hubPlcRetryIntevalSec = 3
)

// Hub-specific registers
//
//go:generate stringer -type=hubRegister
type hubRegister int

const (
	hubIoConnection hubRegister = iota
	hubLightRed
	hubLightGreen
	hubLightBlue
	hubCount
	hubLedAnimation
	hubRegisterCount
)

// Hub-specific coils
//
//go:generate stringer -type=hubCoil
type hubCoil int

const (
	hubHeartbeat hubCoil = iota
	hubActive
	hubBall1Fault
	hubBall2Fault
	hubBall3Fault
	hubBall4Fault
	hubBall1
	hubBall2
	hubBall3
	hubBall4
	hubCoilCount
)

// Hub LED animations
//
//go:generate stringer -type=LedAnimation
type LedAnimation int

const (
	AnimationRgbControl       LedAnimation = 0
	AnimationBluePurpleFading LedAnimation = 1
	AnimationRedSolid         LedAnimation = 2
	AnimationBlueSolid        LedAnimation = 3
	AnimationRedPulsing       LedAnimation = 4
	AnimationBluePulsing      LedAnimation = 5
	AnimationRedChasers       LedAnimation = 6
	AnimationBlueChasers      LedAnimation = 7
)

// GetAnimationNames returns a slice of all available animation names.
func GetAnimationNames() []string {
	return []string{
		AnimationRgbControl.String(),
		AnimationBluePurpleFading.String(),
		AnimationRedSolid.String(),
		AnimationBlueSolid.String(),
		AnimationRedPulsing.String(),
		AnimationBluePulsing.String(),
		AnimationRedChasers.String(),
		AnimationBlueChasers.String(),
	}
}

// NewHubPlc creates a new ModbusHubPlc instance for the given alliance.
func NewHubPlc(allianceName string) *ModbusHubPlc {
	plc := new(ModbusHubPlc)
	plc.allianceName = allianceName
	return plc
}

func (plc *ModbusHubPlc) SetAddress(address string) {
	if address != plc.address {
		plc.address = address
		plc.resetConnection()
	}

	if plc.ioChangeNotifier == nil {
		// Register a notifier that listeners can subscribe to to get websocket updates about I/O value changes.
		notifierName := fmt.Sprintf("%sHubPlcIoChange", plc.allianceName)
		plc.ioChangeNotifier = websocket.NewNotifier(notifierName, plc.generateIoChangeMessage)
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

// SetHubCount sends a command to the PLC to set the hub count to the specified value.
func (plc *ModbusHubPlc) SetHubCount(count int) {
	if plc.isHealthy {
		err := plc.sendSetHubCountCommand(uint16(count))
		if err != nil {
			log.Printf("Error setting hub count: %v", err)
		}
	}
}

// ResetHubCount sends a reset command to the PLC to set the hub count to 0.
func (plc *ModbusHubPlc) ResetHubCount() {
	plc.SetHubCount(0)
}

// SetHubAnimation sets the LED animation mode.
// Possible values: AnimationRgbControl, AnimationBluePurpleFading, AnimationRedSolid, AnimationBlueSolid,
// AnimationRedPulsing, AnimationBluePulsing, AnimationRedChasers, AnimationBlueChasers
func (plc *ModbusHubPlc) SetHubAnimation(animation LedAnimation) {
	plc.registers[hubLedAnimation] = uint16(animation)
}

// SetHubLight sets the state of the hub LED light using animation mode AnimationRgbControl (RGB control).
func (plc *ModbusHubPlc) SetHubLight(active bool) {
	plc.SetHubAnimation(AnimationRgbControl)
	if active {
		switch plc.allianceName {
		case "red":
			plc.SetHubLightColor(255, 0, 0)
		case "blue":
			plc.SetHubLightColor(0, 0, 255)
		default:
			plc.SetHubLightColor(255, 255, 255)
		}
	} else {
		// Turn off all lights
		plc.SetHubLightColor(0, 0, 0)
	}
}

// SetHubLightColor sets the hub LED light to a specific RGB color using AnimationRgbControl.
func (plc *ModbusHubPlc) SetHubLightColor(red, green, blue uint16) {
	plc.registers[hubLedAnimation] = uint16(AnimationRgbControl) // Set animation to RGB control
	plc.registers[hubLightRed] = red
	plc.registers[hubLightGreen] = green
	plc.registers[hubLightBlue] = blue
}

// SetHubChaserAnimation sets the appropriate white chaser animation for this hub's alliance.
// This is used during the transition period to indicate an alliance is about to become inactive.
func (plc *ModbusHubPlc) SetHubChaserAnimation() {
	if plc.allianceName == "red" {
		plc.SetHubAnimation(AnimationRedChasers) // Red with white chasers
	} else if plc.allianceName == "blue" {
		plc.SetHubAnimation(AnimationBlueChasers) // Blue with white chasers
	}
}

// SetHubActive sets the active state for the hub and controls the LED light.
func (plc *ModbusHubPlc) SetHubActive(active bool) {
	plc.coils[hubActive] = active
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
	plc.modbusLock.Lock()
	defer plc.modbusLock.Unlock()

	plc.isHealthy = plc.readRegisters() && plc.readCoils() && plc.writeRegisters() && plc.writeCoils()
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

func (plc *ModbusHubPlc) readCoils() bool {
	if plc.client == nil {
		return false
	}

	coils, err := plc.client.ReadCoils(0, uint16(hubCoilCount))
	if err != nil {
		log.Printf("%s hub PLC error reading coils: %v", plc.allianceName, err)
		return false
	}

	// Only read input coils (faults and ball states), preserving output coils (heartbeat and active)
	// Coils are packed as bits in bytes, so we need to unpack them
	for i := hubBall1Fault; i < hubCoilCount; i++ {
		byteIndex := int(i) / 8
		bitIndex := int(i) % 8
		if byteIndex < len(coils) {
			plc.coils[i] = (coils[byteIndex] & (1 << uint(bitIndex))) != 0
		}
	}

	return true
}

func (plc *ModbusHubPlc) writeRegisters() bool {
	if plc.client == nil {
		return false
	}

	// Write the light color registers (hubLightRed, hubLightGreen, hubLightBlue)
	lightBytes := make([]byte, 6) // 3 registers * 2 bytes each
	lightBytes[0] = byte(plc.registers[hubLightRed] >> 8)
	lightBytes[1] = byte(plc.registers[hubLightRed])
	lightBytes[2] = byte(plc.registers[hubLightGreen] >> 8)
	lightBytes[3] = byte(plc.registers[hubLightGreen])
	lightBytes[4] = byte(plc.registers[hubLightBlue] >> 8)
	lightBytes[5] = byte(plc.registers[hubLightBlue])

	_, err := plc.client.WriteMultipleRegisters(uint16(hubLightRed), 3, lightBytes)
	if err != nil {
		log.Printf("%s hub PLC error writing registers: %v", plc.allianceName, err)
		return false
	}

	// Write the animation register separately
	animationBytes := make([]byte, 2)
	animationBytes[0] = byte(plc.registers[hubLedAnimation] >> 8)
	animationBytes[1] = byte(plc.registers[hubLedAnimation])

	_, err = plc.client.WriteMultipleRegisters(uint16(hubLedAnimation), 1, animationBytes)
	if err != nil {
		log.Printf("%s hub PLC error writing animation register: %v", plc.allianceName, err)
		return false
	}
	return true
}

// sendSetHubCountCommand sends a command to the PLC to set the hub count to the specified value.
func (plc *ModbusHubPlc) sendSetHubCountCommand(count uint16) error {
	plc.modbusLock.Lock()
	defer plc.modbusLock.Unlock()

	if plc.client == nil {
		return fmt.Errorf("PLC client is not connected")
	}

	countBytes := make([]byte, 2)
	countBytes[0] = byte(count >> 8)
	countBytes[1] = byte(count)

	_, err := plc.client.WriteMultipleRegisters(uint16(hubCount), 1, countBytes)
	if err != nil {
		return fmt.Errorf("%s hub PLC error setting hub count: %v", plc.allianceName, err)
	}
	return nil
}

func (plc *ModbusHubPlc) writeCoils() bool {
	if plc.client == nil {
		return false
	}

	// Send a heartbeat to the PLC so that it can disable outputs if the connection is lost.
	plc.coils[hubHeartbeat] = true

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

func (plc *ModbusHubPlc) generateIoChangeMessage() any {
	return &struct {
		Registers []uint16
		Coils     []bool
	}{plc.registers[:], plc.coils[:]}
}
