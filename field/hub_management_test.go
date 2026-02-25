// Copyright 2026 Team 254. All Rights Reserved.
//
// Tests for 2026 REBUILT HUB status management based on match phase and AUTO fuel scoring.

package field

import (
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
)

func TestHubStatusAutoPhase(t *testing.T) {
	arena := setupTestArena(t)
	arena.LoadTestMatch()

	// During AUTO, both hubs should be active
	arena.MatchState = AutoPeriod
	arena.updateHubStatus(0)

	assert.True(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive)
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive)
}

func TestHubStatusTransitionShift(t *testing.T) {
	arena := setupTestArena(t)
	arena.LoadTestMatch()

	// During TRANSITION SHIFT (2:20-2:10), both hubs should be active
	// Match timing: Warmup=3, Auto=20, Pause=3, Teleop=140, Warning=30
	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.AutoDurationSec = 20
	game.MatchTiming.PauseDurationSec = 3
	game.MatchTiming.TeleopDurationSec = 140
	game.MatchTiming.WarningRemainingDurationSec = 30
	game.MatchTiming.TransitionDurationSec = 10

	arena.MatchState = TeleopPeriod
	// transitionShiftStart = warmup(3) + auto(20) + pause(3) = 26 seconds
	// transitionShiftEnd = 26 + 10 = 36 seconds
	// Test at matchTimeSec = 30 (in transition shift)
	arena.updateHubStatus(30)

	assert.True(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive)
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive)
}

func TestHubStatusAllianceShifts(t *testing.T) {
	arena := setupTestArena(t)
	arena.LoadTestMatch()

	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.AutoDurationSec = 20
	game.MatchTiming.PauseDurationSec = 2
	game.MatchTiming.TeleopDurationSec = 140
	game.MatchTiming.WarningRemainingDurationSec = 30
	game.MatchTiming.TransitionDurationSec = 10
	game.MatchTiming.AllianceShiftDurationSec = 25

	// Set up AUTO scoring: Red wins with 10 fuel vs Blue's 5 fuel
	arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel = 10
	arena.BlueRealtimeScore.CurrentScore.Hub.AutoFuel = 5

	arena.MatchState = TeleopPeriod
	arena.autoWinnerDetermined = true
	arena.autoWinningAlliance = "red"

	// Total match time: warmup(3) + auto(20) + pause(2) + teleop(140) = 165 seconds
	// transitionShiftStart = 3 + 20 + 2 = 25 seconds (start of teleop)
	// transitionShiftEnd = 25 + 10 = 35 seconds
	// endGameStart = 3 + 20 + 2 + 140 - 30 = 135 seconds

	// matchTimeSec now counts UP from 0
	// SHIFT 1: matchTimeSec 35-60 (25 seconds), shiftNumber = 0
	// SHIFT 2: matchTimeSec 60-85 (25 seconds), shiftNumber = 1
	// SHIFT 3: matchTimeSec 85-110 (25 seconds), shiftNumber = 2
	// SHIFT 4: matchTimeSec 110-135 (25 seconds), shiftNumber = 3
	// END GAME: matchTimeSec 135+

	// SHIFT 1: shiftNumber = 0
	// For RED winning: redActive = (0 % 2 == 1) = false -> Red INACTIVE, Blue ACTIVE ✓
	arena.updateHubStatus(40) // matchTimeSec = 40, in SHIFT 1
	assert.False(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be inactive in SHIFT 1 (shift 0)")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be active in SHIFT 1 (shift 0)")

	// SHIFT 2: shiftNumber = 1
	// For RED winning: redActive = (1 % 2 == 1) = true -> Red ACTIVE, Blue INACTIVE ✓
	arena.updateHubStatus(65) // matchTimeSec = 65, in SHIFT 2
	assert.True(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be active in SHIFT 2 (shift 1)")
	assert.False(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be inactive in SHIFT 2 (shift 1)")

	// SHIFT 3: shiftNumber = 2
	// For RED winning: redActive = (2 % 2 == 1) = false -> Red INACTIVE, Blue ACTIVE ✓
	arena.updateHubStatus(90) // matchTimeSec = 90, in SHIFT 3
	assert.False(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be inactive in SHIFT 3 (shift 2)")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be active in SHIFT 3 (shift 2)")

	// SHIFT 4: shiftNumber = 3
	// For RED winning: redActive = (3 % 2 == 1) = true -> Red ACTIVE, Blue INACTIVE ✓
	arena.updateHubStatus(115) // matchTimeSec = 115, in SHIFT 4
	assert.True(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be active in SHIFT 4 (shift 3)")
	assert.False(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be inactive in SHIFT 4 (shift 3)")
}

func TestHubStatusEndGame(t *testing.T) {
	arena := setupTestArena(t)
	arena.LoadTestMatch()

	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.AutoDurationSec = 20
	game.MatchTiming.PauseDurationSec = 2
	game.MatchTiming.TeleopDurationSec = 140
	game.MatchTiming.WarningRemainingDurationSec = 30

	arena.MatchState = TeleopPeriod
	arena.autoWinnerDetermined = true
	arena.autoWinningAlliance = "red"

	// During END GAME (last 30 seconds), both hubs should be active
	// Total: 165 seconds, EndGame starts at 135 seconds from start (warmup+auto+pause+teleop-warning)
	// matchTimeSec = 140 (in end game period)
	arena.updateHubStatus(140)

	assert.True(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be active in END GAME")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be active in END GAME")
}

func TestHubStatusAutoWinnerDetermined(t *testing.T) {
	arena := setupTestArena(t)
	arena.LoadTestMatch()

	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.AutoDurationSec = 20
	game.MatchTiming.PauseDurationSec = 2
	game.MatchTiming.TeleopDurationSec = 140
	game.MatchTiming.WarningRemainingDurationSec = 30

	// Set up AUTO scoring: Blue wins
	arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel = 5
	arena.BlueRealtimeScore.CurrentScore.Hub.AutoFuel = 15

	arena.MatchState = TeleopPeriod

	// First call should determine winner
	assert.False(t, arena.autoWinnerDetermined)
	// Call with matchTimeSec in transition shift (e.g., 27 seconds from start)
	arena.updateHubStatus(27)
	assert.True(t, arena.autoWinnerDetermined)
	assert.Equal(t, "blue", arena.autoWinningAlliance)
}

func TestHubStatusTiedAutoFuel(t *testing.T) {
	arena := setupTestArena(t)
	arena.LoadTestMatch()

	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.AutoDurationSec = 20
	game.MatchTiming.PauseDurationSec = 2
	game.MatchTiming.TeleopDurationSec = 140
	game.MatchTiming.WarningRemainingDurationSec = 30

	// Set up tied AUTO scoring
	arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel = 10
	arena.BlueRealtimeScore.CurrentScore.Hub.AutoFuel = 10

	arena.MatchState = TeleopPeriod

	// Should randomly select red or blue when tied once alliance shifts begin.
	// transitionShiftStart = 3 + 20 + 2 = 25
	// transitionShiftEnd = 25 + 10 = 35
	arena.updateHubStatus(36)
	// Just verify one was selected (can't test randomness reliably)
	assert.True(t, arena.autoWinningAlliance == "red" || arena.autoWinningAlliance == "blue")
}
