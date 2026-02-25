// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	Hub             Hub
	Fouls           []Foul
	RobotsBypassed  [3]bool
	EndgameStatuses [3]EndgameStatus
	AutoStatuses    [3]EndgameStatus
	PlayoffDq       bool
}

// Game-specific settings that can be changed via the settings.
var EnergizedFuelThreshold = 100
var SuperchargedFuelThreshold = 360
var TraversalTowerThreshold = 50

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameL1
	EndgameL2
	EndgameL3
)

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// If this alliance is DQ'd in playoffs, zero out the score.
	if score.PlayoffDq {
		return summary
	}

	// Calculate AUTO Tower points.
	for _, status := range score.AutoStatuses {
		switch status {
		case EndgameL1:
			summary.AutoTowerPoints += 15
		default:
		}
	}

	// Calculate FUEL points (1 point per FUEL, only if HUB is active).
	summary.AutoFuelPoints = score.Hub.AutoFuelPoints()
	summary.AutoPoints = summary.AutoTowerPoints + summary.AutoFuelPoints
	// NumFuel is total fuel scored (for display), but only active fuel counts for points
	summary.NumFuel = score.Hub.TotalFuel()
	teleopFuelPoints := score.Hub.TeleopFuelPoints()

	// Calculate endgame points.
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgameL1:
			summary.TowerPoints += 10
		case EndgameL2:
			summary.TowerPoints += 20
		case EndgameL3:
			summary.TowerPoints += 30
		default:
		}
	}

	// Match points = FUEL points (auto + teleop) + TOWER points.
	summary.TowerPoints += summary.AutoTowerPoints
	summary.MatchPoints = summary.AutoFuelPoints + teleopFuelPoints + summary.TowerPoints

	// ActiveFuelPoints represents the total fuel points (AUTO + TELEOP) that count
	// toward ranking point thresholds.
	summary.ActiveFuelPoints = summary.AutoFuelPoints + teleopFuelPoints

	// Calculate penalty points.
	if opponentScore != nil {
		for _, foul := range opponentScore.Fouls {
			summary.FoulPoints += foul.PointValue()
			// Store the number of major fouls since it is used to break ties in playoffs.
			if foul.IsMajor {
				summary.NumOpponentMajorFouls++
			}
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate ranking points (0-3 for win/tie + 0-3 for bonus RPs).
	// Bonus RPs are: ENERGIZED RP, SUPERCHARGED RP, TRAVERSAL RP.

	// ENERGIZED RP - FUEL at or above threshold.
	if score.Hub.IsEnergized(EnergizedFuelThreshold) {
		summary.EnergizedRankingPoint = true
	}

	// SUPERCHARGED RP - FUEL at or above higher threshold.
	if score.Hub.IsSupercharged(SuperchargedFuelThreshold) {
		summary.SuperchargedRankingPoint = true
	}

	// TRAVERSAL RP - TOWER points at or above threshold.
	if summary.TowerPoints >= TraversalTowerThreshold {
		summary.TraversalRankingPoint = true
	}

	// Count bonus ranking points.
	if summary.EnergizedRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.SuperchargedRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.TraversalRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Equals returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.Hub != other.Hub ||
		score.RobotsBypassed != other.RobotsBypassed ||
		score.PlayoffDq != other.PlayoffDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
