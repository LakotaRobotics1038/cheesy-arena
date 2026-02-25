// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Tests for the 2026 REBUILT Hub element.

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	// Hub fuel accessors return the raw counts regardless of the IsActive
	// flag; tests should reflect that behavior.
	assert.Equal(t, 5, hub.AutoFuelPoints())
	assert.Equal(t, 8, hub.TeleopFuelPoints())
}

func TestHubTotalFuel(t *testing.T) {
	hub := &Hub{
		AutoFuel:   3,
		TeleopFuel: 7,
	}

	assert.Equal(t, 10, hub.TotalFuel())
}

func TestHubActivateDeactivate(t *testing.T) {
	hub := NewHub()
	assert.False(t, hub.IsActive)

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
	assert.NotEqual(t, originalTime, hub.LastLEDToggle)
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
