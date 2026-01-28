// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	Hub            Hub
	Fouls          []Foul
	RobotsBypassed [3]bool
	PlayoffDq      bool
}

// Game-specific settings that can be changed via the settings.
var EnergizedFuelThreshold = 100
var SuperchargedFuelThreshold = 360
var TraversalTowerThreshold = 50

// Summarize calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// If this alliance is DQ'd in playoffs, zero out the score.
	if score.PlayoffDq {
		return summary
	}

	// Calculate FUEL points (1 point per FUEL, only if HUB is active).
	summary.FuelPoints = score.Hub.AutoFuelPoints() + score.Hub.TeleopFuelPoints()
	summary.AutoFuelPoints = score.Hub.AutoFuelPoints()
	summary.NumFuel = score.Hub.TotalFuel()

	// Calculate TOWER points.
	summary.TowerPoints = score.Hub.AutoTowerPoints() + score.Hub.TeleopTowerPoints()
	// Match points = FUEL points + TOWER points.
	summary.MatchPoints = summary.FuelPoints + summary.TowerPoints

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
	if score.Hub.MeetsTowerThreshold(TraversalTowerThreshold) {
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
