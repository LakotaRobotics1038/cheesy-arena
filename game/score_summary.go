// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the calculated totals of a match score.

package game

type ScoreSummary struct {
	FuelPoints             int
	NumFuel                int
	AutoFuelPoints         int
	TowerPoints            int
	MatchPoints            int
	FoulPoints             int
	Score                  int
	EnergizedRankingPoint  bool
	SuperchargedRankingPoint bool
	TraversalRankingPoint  bool
	BonusRankingPoints     int
	NumOpponentMajorFouls  int
}

type MatchStatus int

const (
	MatchScheduled MatchStatus = iota
	MatchHidden
	RedWonMatch
	BlueWonMatch
	TieMatch
)

func (t MatchStatus) Get() MatchStatus {
	return t
}

// Determines the winner of the match given the score summaries for both alliances.
func DetermineMatchStatus(redScoreSummary, blueScoreSummary *ScoreSummary, applyPlayoffTiebreakers bool) MatchStatus {
	if status := comparePoints(redScoreSummary.Score, blueScoreSummary.Score); status != TieMatch {
		return status
	}

	if applyPlayoffTiebreakers {
		// Check scoring breakdowns to resolve playoff ties (2026 playoff tiebreaker rules).
		// 1. Fewest opponent major fouls
		if status := comparePoints(
			redScoreSummary.NumOpponentMajorFouls, blueScoreSummary.NumOpponentMajorFouls,
		); status != TieMatch {
			return status
		}
		// 2. Most AUTO fuel points
		if status := comparePoints(redScoreSummary.AutoFuelPoints, blueScoreSummary.AutoFuelPoints); status != TieMatch {
			return status
		}
		// 3. Most tower points
		if status := comparePoints(redScoreSummary.TowerPoints, blueScoreSummary.TowerPoints); status != TieMatch {
			return status
		}
	}

	return TieMatch
}

// Helper method to compare the red and blue alliance point totals and return the appropriate MatchStatus.
func comparePoints(redPoints, bluePoints int) MatchStatus {
	if redPoints > bluePoints {
		return RedWonMatch
	}
	if redPoints < bluePoints {
		return BlueWonMatch
	}
	return TieMatch
}
