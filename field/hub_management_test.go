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
	// Match timing: Warmup=3, Auto=20, Pause=2, Teleop=140, Warning=30
	game.MatchTiming.WarmupDurationSec = 3
	game.MatchTiming.AutoDurationSec = 20
	game.MatchTiming.PauseDurationSec = 2
	game.MatchTiming.TeleopDurationSec = 140
	game.MatchTiming.WarningRemainingDurationSec = 30

	arena.MatchState = TeleopPeriod
	// At 155 seconds (start of TRANSITION SHIFT: warmup 3 + auto 20 + 10)
	arena.updateHubStatus(155)

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

	// Set up AUTO scoring: Red wins with 10 fuel vs Blue's 5 fuel
	arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel = 10
	arena.BlueRealtimeScore.CurrentScore.Hub.AutoFuel = 5

	arena.MatchState = TeleopPeriod
	arena.autoWinnerDetermined = true
	arena.autoWinningAlliance = "red"

	// Total match time: warmup(3) + auto(20) + pause(2) + teleop(140) = 165 seconds
	// transitionShiftStart = 3 + 20 + 10 = 33 seconds
	// endGameStart = 165 - 30 = 135 seconds

	// TRANSITION SHIFT ends at 33 seconds, so when matchTimeSec = 165 - 33 = 132
	// SHIFT 1 runs from 33-58 seconds (25 seconds), matchTimeSec = 132-107
	// SHIFT 2 runs from 58-83 seconds (25 seconds), matchTimeSec = 107-82
	// SHIFT 3 runs from 83-108 seconds (25 seconds), matchTimeSec = 82-57
	// SHIFT 4 runs from 108-133 seconds (25 seconds), matchTimeSec = 57-32
	// END GAME starts at 135 seconds from start... wait that's before shifts end

	// Let me recalculate endGameStart:
	// endGameStart = warmup(3) + auto(20) + pause(2) + teleop(140) - warning(30)
	// endGameStart = 165 - 30 = 135 seconds
	// But transition shift is at 33, so endgame is AFTER all shifts
	// Actually the formulation is wrong. endGameStart should be when warning period starts
	// which is: total_time - warning_time = 135 seconds after warmup+auto+pause starts
	// NO: endGameStart = warmup+auto+pause + teleop - warning = 3+20+2+140-30 = 135
	// But teleop = 140 and only lasts 140 seconds, so warning starts at 140-30 = 110 seconds into teleop
	// which is 3+20+2+110 = 135 seconds into match

	// So END GAME (warning period) starts at 135 seconds
	// SHIFT 1: elapsedTime 33-58, matchTimeSec 132-107
	// SHIFT 2: elapsedTime 58-83, matchTimeSec 107-82
	// SHIFT 3: elapsedTime 83-108, matchTimeSec 82-57
	// SHIFT 4: elapsedTime 108-133, matchTimeSec 57-32
	// END GAME: elapsedTime 135+, matchTimeSec 30-

	// SHIFT 1: secIntoShifts = elapsedTime - 33, shiftNumber = 0
	// For RED winning: redActive = (0 % 2 == 1) = false -> Red INACTIVE, Blue ACTIVE ✓
	arena.updateHubStatus(130) // elapsedTime ≈ 35
	assert.False(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be inactive in SHIFT 1 (shift 0)")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be active in SHIFT 1 (shift 0)")

	// SHIFT 2: secIntoShifts = 58 - 33 = 25, shiftNumber = 1
	// For RED winning: redActive = (1 % 2 == 1) = true -> Red ACTIVE, Blue INACTIVE ✓
	arena.updateHubStatus(105) // elapsedTime ≈ 60
	assert.True(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be active in SHIFT 2 (shift 1)")
	assert.False(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be inactive in SHIFT 2 (shift 1)")

	// SHIFT 3: secIntoShifts = 83 - 33 = 50, shiftNumber = 2
	// For RED winning: redActive = (2 % 2 == 1) = false -> Red INACTIVE, Blue ACTIVE ✓
	arena.updateHubStatus(80) // elapsedTime ≈ 85
	assert.False(t, arena.RedRealtimeScore.CurrentScore.Hub.IsActive, "Red should be inactive in SHIFT 3 (shift 2)")
	assert.True(t, arena.BlueRealtimeScore.CurrentScore.Hub.IsActive, "Blue should be active in SHIFT 3 (shift 2)")

	// SHIFT 4: secIntoShifts = 108 - 33 = 75, shiftNumber = 3
	// For RED winning: redActive = (3 % 2 == 1) = true -> Red ACTIVE, Blue INACTIVE ✓
	arena.updateHubStatus(55) // elapsedTime ≈ 110
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
	// Total: 165 seconds, EndGame starts at 30 second countdown (165 - 30 = 135)
	arena.updateHubStatus(25)

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
	arena.updateHubStatus(150)
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

	// Should default to red when tied
	arena.updateHubStatus(150)
	assert.Equal(t, "red", arena.autoWinningAlliance)
}
