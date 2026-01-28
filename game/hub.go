// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Scoring logic for the 2026 REBUILT game with FUEL and TOWER scoring.

package game

import "time"

type Hub struct {
	AutoFuel         int        // FUEL scored during AUTO period
	TeleopFuel       int        // FUEL scored during TELEOP period
	AutoTowerLevel   TowerLevel // Tower level achieved during AUTO
	TeleopTowerLevel TowerLevel // Tower level achieved during TELEOP
	IsActive         bool           // Whether the HUB is active for FUEL scoring
	LEDState         bool           // Current LED state
	LastLEDToggle    time.Time      // Last time LED state was toggled
	LEDCyclePeriod   time.Duration  // Period for LED cycling
	DeactivationTime time.Time      // When the HUB will be deactivated (for warning flash)
}

// TowerLevel represents the climbing level a robot achieves on the TOWER.
type TowerLevel int

const (
	TowerLevelNone TowerLevel = iota
	TowerLevel1
	TowerLevel2
	TowerLevel3
)

// NewHub creates a new Hub with default LED cycling period.
func NewHub() *Hub {
	return &Hub{
		LEDCyclePeriod: 5 * time.Second,
		LEDState:       false,
		LastLEDToggle:  time.Now(),
		IsActive:       true,
	}
}

// UpdateLEDState updates the LED state based on the configured cycle period.
// If within 3 seconds of deactivation, flashes rapidly as a warning.
func (hub *Hub) UpdateLEDState(currentTime time.Time) {
	cycleFreq := hub.LEDCyclePeriod

	// If within 3 seconds of deactivation, flash faster as a warning
	if !hub.DeactivationTime.IsZero() &&
	   currentTime.Add(3*time.Second).After(hub.DeactivationTime) {
		cycleFreq = 500 * time.Millisecond // Rapid warning flash
	}

	if currentTime.Sub(hub.LastLEDToggle) >= cycleFreq {
		hub.LEDState = !hub.LEDState
		hub.LastLEDToggle = currentTime
	}
}

// Deactivate turns off the HUB for FUEL scoring (e.g., at the end of a period or match).
func (hub *Hub) Deactivate() {
	hub.IsActive = false
}

// Activate turns on the HUB for FUEL scoring.
func (hub *Hub) Activate() {
	hub.IsActive = true
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

// AutoFuelPoints calculates points from FUEL scored during AUTO. Only counts if HUB is active.
func (hub *Hub) AutoFuelPoints() int {
	if hub.IsActive {
		return hub.AutoFuel
	}
	return 0
}

// TeleopFuelPoints calculates points from FUEL scored during TELEOP. Only counts if HUB is active.
func (hub *Hub) TeleopFuelPoints() int {
	if hub.IsActive {
		return hub.TeleopFuel
	}
	return 0
}

// TotalFuel returns the total FUEL scored across AUTO and TELEOP.
func (hub *Hub) TotalFuel() int {
	return hub.AutoFuel + hub.TeleopFuel
}

// AutoTowerPoints calculates points from TOWER climbing during AUTO.
// Level 1 = 15 points, max 2 robots per alliance per period.
// Higher levels only in TELEOP.
func (hub *Hub) AutoTowerPoints() int {
	if hub.AutoTowerLevel == TowerLevel1 {
		return 15
	}
	return 0
}

// TeleopTowerPoints calculates points from TOWER climbing during TELEOP.
// Level 1 = 10 points, Level 2 = 20 points, Level 3 = 30 points.
func (hub *Hub) TeleopTowerPoints() int {
	switch hub.TeleopTowerLevel {
	case TowerLevel1:
		return 10
	case TowerLevel2:
		return 20
	case TowerLevel3:
		return 30
	default:
		return 0
	}
}

// TotalTowerPoints returns the total TOWER points from both AUTO and TELEOP.
func (hub *Hub) TotalTowerPoints() int {
	return hub.AutoTowerPoints() + hub.TeleopTowerPoints()
}

// HasAutoTowerClimb returns true if any robot climbed during AUTO.
func (hub *Hub) HasAutoTowerClimb() bool {
	return hub.AutoTowerLevel != TowerLevelNone
}

// HasTeleopTowerClimb returns true if any robot climbed during TELEOP.
func (hub *Hub) HasTeleopTowerClimb() bool {
	return hub.TeleopTowerLevel != TowerLevelNone
}

// IsEnergized returns true if FUEL meets the ENERGIZED threshold.
func (hub *Hub) IsEnergized(threshold int) bool {
	return hub.TotalFuel() >= threshold
}

// IsSupercharged returns true if FUEL meets the SUPERCHARGED threshold.
func (hub *Hub) IsSupercharged(threshold int) bool {
	return hub.TotalFuel() >= threshold
}

// MeetsTowerThreshold returns true if TOWER points meet the given threshold.
func (hub *Hub) MeetsTowerThreshold(threshold int) bool {
	return hub.TotalTowerPoints() >= threshold
}
