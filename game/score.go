// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	RedHub     Hub
	BlueHub    Hub
	Fouls      []Foul
	PlayoffDq  bool
}

// Game-specific settings that can be changed via the settings.
var EnergizedFuelThreshold = 100
var SuperchargedFuelThreshold = 360
var TraversalTowerThreshold = 50

// Summarize calculates and returns the summary fields used for ranking and display.
// allianceColor should be "red" or "blue"
func (score *Score) Summarize(allianceColor string, opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	var ownHub *Hub

	if allianceColor == "red" {
		ownHub = &score.RedHub
	} else {
		ownHub = &score.BlueHub
	}

	// Calculate FUEL points (1 point per FUEL, only if HUB is active).
	summary.FuelPoints = ownHub.AutoFuelPoints() + ownHub.TeleopFuelPoints()
	summary.NumFuel = ownHub.TotalFuel()

	// Calculate TOWER points.
	summary.TowerPoints = ownHub.AutoTowerPoints() + ownHub.TeleopTowerPoints()

	// Match points = FUEL points + TOWER points.
	summary.MatchPoints = summary.FuelPoints + summary.TowerPoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of major fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate ranking points (0-3 for win/tie + 0-3 for bonus RPs).
	// Bonus RPs are: ENERGIZED RP, SUPERCHARGED RP, TRAVERSAL RP.

	// ENERGIZED RP - FUEL at or above threshold.
	if ownHub.IsEnergized(EnergizedFuelThreshold) {
		summary.EnergizedRankingPoint = true
	}

	// SUPERCHARGED RP - FUEL at or above higher threshold.
	if ownHub.IsSupercharged(SuperchargedFuelThreshold) {
		summary.SuperchargedRankingPoint = true
	}

	// TRAVERSAL RP - TOWER points at or above threshold.
	if ownHub.MeetsTowerThreshold(TraversalTowerThreshold) {
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
	if score.RedHub != other.RedHub ||
		score.BlueHub != other.BlueHub ||
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
