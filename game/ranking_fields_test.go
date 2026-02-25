// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sort"
	"testing"
)

func TestAddScoreSummary(t *testing.T) {
	rand.Seed(0)
	redSummary := &ScoreSummary{
		NumFuel:                  45,
		TowerPoints:              30,
		MatchPoints:              75,
		Score:                    75,
		EnergizedRankingPoint:    true,
		SuperchargedRankingPoint: false,
		TraversalRankingPoint:    true,
		BonusRankingPoints:       2,
	}
	blueSummary := &ScoreSummary{
		NumFuel:                  20,
		TowerPoints:              40,
		MatchPoints:              60,
		Score:                    60,
		EnergizedRankingPoint:    false,
		SuperchargedRankingPoint: true,
		TraversalRankingPoint:    false,
		BonusRankingPoints:       1,
	}
	rankingFields := RankingFields{}

	// Add a win (red 75 > blue 60). Red wins 3 RP + 2 bonus = 5 RP.
	rankingFields.AddScoreSummary(redSummary, blueSummary, false)
	assert.Equal(t, RankingFields{5, 75, 45, 30, 0.9451961492941164, 1, 0, 0, 0, 1}, rankingFields)

	// Add another match (blue 60 < red 75, so blue loses 0 RP + 1 bonus = 1 RP).
	rankingFields.AddScoreSummary(blueSummary, redSummary, false)
	assert.Equal(t, RankingFields{6, 135, 65, 70, 0.24496508529377975, 1, 1, 0, 0, 2}, rankingFields)

	// Add a tie (red 75 == red 75). Red ties 1 RP + 2 bonus = 3 RP.
	rankingFields.AddScoreSummary(redSummary, redSummary, false)
	assert.Equal(t, RankingFields{9, 210, 110, 100, 0.6559562651954052, 1, 1, 1, 0, 3}, rankingFields)

	// Add a disqualification (no points awarded).
	rankingFields.AddScoreSummary(blueSummary, redSummary, true)
	assert.Equal(t, RankingFields{9, 210, 110, 100, 0.05434383959970039, 1, 1, 1, 1, 4}, rankingFields)
}

func TestSortRankings(t *testing.T) {
	// Verify that ranking sort works correctly by checking that higher ranking points come first
	rankings := make(Rankings, 2)
	rankings[0] = Ranking{1, 0, 0, RankingFields{50, 50, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	rankings[1] = Ranking{2, 0, 0, RankingFields{51, 50, 50, 50, 0.50, 3, 2, 1, 0, 10}}
	sort.Sort(rankings)
	// Team 2 has higher ranking points (51 vs 50), so should come first
	assert.Equal(t, 2, rankings[0].TeamId)
	assert.Equal(t, 1, rankings[1].TeamId)
}
