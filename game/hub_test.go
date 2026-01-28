// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Tests for the 2026 REBUILT Hub element.

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHubLEDCycling(t *testing.T) {
	hub := NewHub()
	assert.False(t, hub.LEDState)

	// Simulate time progression
	startTime := time.Now()
	hub.LastLEDToggle = startTime

	// Before cycle period, LED state should not change
	hub.UpdateLEDState(startTime.Add(3 * time.Second))
	assert.False(t, hub.LEDState)

	// After cycle period, LED state should toggle
	hub.UpdateLEDState(startTime.Add(6 * time.Second))
	assert.True(t, hub.LEDState)

	// Another cycle should toggle it back
	hub.UpdateLEDState(startTime.Add(12 * time.Second))
	assert.False(t, hub.LEDState)
}

func TestHubLEDCyclingCustomPeriod(t *testing.T) {
	hub := NewHub()
	hub.LEDCyclePeriod = 2 * time.Second
	assert.False(t, hub.LEDState)

	startTime := time.Now()
	hub.LastLEDToggle = startTime

	// After custom cycle period
	hub.UpdateLEDState(startTime.Add(2 * time.Second))
	assert.True(t, hub.LEDState)

	hub.UpdateLEDState(startTime.Add(4 * time.Second))
	assert.False(t, hub.LEDState)
}

func TestHubFuelPoints(t *testing.T) {
	hub := &Hub{
		AutoFuel:   5,
		TeleopFuel: 8,
		IsActive:   true,
	}

	assert.Equal(t, 5, hub.AutoFuelPoints())
	assert.Equal(t, 8, hub.TeleopFuelPoints())
	assert.Equal(t, 13, hub.AutoFuelPoints()+hub.TeleopFuelPoints())
}

func TestHubFuelPointsInactive(t *testing.T) {
	hub := &Hub{
		AutoFuel:   5,
		TeleopFuel: 8,
		IsActive:   false,
	}

	assert.Equal(t, 0, hub.AutoFuelPoints())
	assert.Equal(t, 0, hub.TeleopFuelPoints())
}

func TestHubTotalFuel(t *testing.T) {
	hub := &Hub{
		AutoFuel:   3,
		TeleopFuel: 7,
	}

	assert.Equal(t, 10, hub.TotalFuel())
}

func TestHubAutoTowerPoints(t *testing.T) {
	hub := &Hub{AutoTowerLevel: TowerLevel1}
	assert.Equal(t, 15, hub.AutoTowerPoints())

	hub.AutoTowerLevel = TowerLevel2
	assert.Equal(t, 0, hub.AutoTowerPoints())

	hub.AutoTowerLevel = TowerLevelNone
	assert.Equal(t, 0, hub.AutoTowerPoints())
}

func TestHubTeleopTowerPoints(t *testing.T) {
	hub := &Hub{TeleopTowerLevel: TowerLevel1}
	assert.Equal(t, 10, hub.TeleopTowerPoints())

	hub.TeleopTowerLevel = TowerLevel2
	assert.Equal(t, 20, hub.TeleopTowerPoints())

	hub.TeleopTowerLevel = TowerLevel3
	assert.Equal(t, 30, hub.TeleopTowerPoints())

	hub.TeleopTowerLevel = TowerLevelNone
	assert.Equal(t, 0, hub.TeleopTowerPoints())
}

func TestHubTotalTowerPoints(t *testing.T) {
	hub := &Hub{
		AutoTowerLevel:   TowerLevel1,
		TeleopTowerLevel: TowerLevel2,
	}

	assert.Equal(t, 35, hub.TotalTowerPoints())
}

func TestHubActivateDeactivate(t *testing.T) {
	hub := NewHub()
	assert.True(t, hub.IsActive)

	hub.Deactivate()
	assert.False(t, hub.IsActive)

	hub.Activate()
	assert.True(t, hub.IsActive)
}

func TestHubToggleLED(t *testing.T) {
	hub := NewHub()
	originalState := hub.LEDState
	originalTime := hub.LastLEDToggle

	hub.ToggleLED()
	assert.Equal(t, !originalState, hub.LEDState)
	assert.True(t, hub.LastLEDToggle.After(originalTime))
}

func TestHubIsEnergized(t *testing.T) {
	hub := &Hub{AutoFuel: 20, TeleopFuel: 15}

	assert.True(t, hub.IsEnergized(30))
	assert.False(t, hub.IsEnergized(40))
	assert.True(t, hub.IsEnergized(35))
}

func TestHubIsSupercharged(t *testing.T) {
	hub := &Hub{AutoFuel: 25, TeleopFuel: 20}

	assert.True(t, hub.IsSupercharged(40))
	assert.False(t, hub.IsSupercharged(50))
	assert.True(t, hub.IsSupercharged(45))
}

func TestHubMeetsTowerThreshold(t *testing.T) {
	hub := &Hub{
		AutoTowerLevel:   TowerLevel1,
		TeleopTowerLevel: TowerLevel2,
	}

	// Total: 15 + 20 = 35
	assert.True(t, hub.MeetsTowerThreshold(30))
	assert.False(t, hub.MeetsTowerThreshold(40))
	assert.True(t, hub.MeetsTowerThreshold(35))
}

func TestHubHasAutoTowerClimb(t *testing.T) {
	hub := &Hub{AutoTowerLevel: TowerLevelNone}
	assert.False(t, hub.HasAutoTowerClimb())

	hub.AutoTowerLevel = TowerLevel1
	assert.True(t, hub.HasAutoTowerClimb())
}

func TestHubHasTeleopTowerClimb(t *testing.T) {
	hub := &Hub{TeleopTowerLevel: TowerLevelNone}
	assert.False(t, hub.HasTeleopTowerClimb())

	hub.TeleopTowerLevel = TowerLevel3
	assert.True(t, hub.HasTeleopTowerClimb())
}
