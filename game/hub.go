// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2026 REBUILT game with FUEL scoring.

package game

import "time"

type Hub struct {
	AutoFuel         int           // FUEL scored during AUTO period
	TeleopFuel       int           // FUEL scored during TELEOP period
	IsActive         bool          // Whether the HUB is active for FUEL scoring
	LEDState         bool          // Current LED state
	LastLEDToggle    time.Time     // Last time LED state was toggled
	LEDCyclePeriod   time.Duration // Period for LED cycling
	DeactivationTime time.Time     // When the HUB will be deactivated (for warning flash)
	ActivationTime   time.Time     // When the HUB was activated
}

// NewHub creates a new Hub with default LED cycling period.
func NewHub() *Hub {
	return &Hub{
		LEDCyclePeriod: 5 * time.Second,
		LEDState:       false,
		LastLEDToggle:  time.Now(),
		IsActive:       false,
	}
}

// UpdateLEDState updates the LED state based on the hub status.
// - Solid ON when hub is active and not near end of shift
// - Flashing when within 3 seconds of deactivation
// - OFF when hub is inactive
func (hub *Hub) UpdateLEDState(currentTime time.Time) {
	// If hub is not active, LED is off
	if !hub.IsActive {
		hub.LEDState = false
		return
	}

	// If within 3 seconds of deactivation, flash as a warning
	if !hub.DeactivationTime.IsZero() &&
		currentTime.Before(hub.DeactivationTime) &&
		hub.DeactivationTime.Sub(currentTime) <= 3*time.Second {
		cycleFreq := 500 * time.Millisecond // Rapid warning flash
		if currentTime.Sub(hub.LastLEDToggle) >= cycleFreq {
			hub.LEDState = !hub.LEDState
			hub.LastLEDToggle = currentTime
		}
	} else {
		// Otherwise, LED is solid ON
		hub.LEDState = true
	}
}

// Deactivate turns off the HUB for FUEL scoring (e.g., at the end of a period or match).
func (hub *Hub) Deactivate() {
	hub.IsActive = false
}

// Activate turns on the HUB for FUEL scoring.
func (hub *Hub) Activate() {
	hub.IsActive = true
	hub.ActivationTime = time.Now()
	// Clear deactivation time when activating
	hub.DeactivationTime = time.Time{}
}

// ToggleLED manually toggles the LED state.
func (hub *Hub) ToggleLED() {
	hub.LEDState = !hub.LEDState
	hub.LastLEDToggle = time.Now()
}

// SetDeactivationTime sets the time when the HUB will become inactive.
// The LED will flash rapidly (500ms) for the 3 seconds before deactivation as a warning.
func (hub *Hub) SetDeactivationTime(t time.Time) {
	hub.DeactivationTime = t
}

// AutoFuelPoints calculates points from FUEL scored during AUTO.
func (hub *Hub) AutoFuelPoints() int {
	return hub.AutoFuel
}

// TeleopFuelPoints calculates points from FUEL scored during TELEOP.
func (hub *Hub) TeleopFuelPoints() int {
	return hub.TeleopFuel
}

// TotalFuel returns the total FUEL scored across AUTO and TELEOP.
func (hub *Hub) TotalFuel() int {
	return hub.AutoFuel + hub.TeleopFuel
}

// IsEnergized returns true if FUEL meets the ENERGIZED threshold.
func (hub *Hub) IsEnergized(threshold int) bool {
	return hub.TotalFuel() >= threshold
}

// IsSupercharged returns true if FUEL meets the SUPERCHARGED threshold.
func (hub *Hub) IsSupercharged(threshold int) bool {
	return hub.TotalFuel() >= threshold
}
