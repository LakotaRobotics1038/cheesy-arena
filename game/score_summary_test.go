// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDetermineMatchStatus(t *testing.T) {
	redScoreSummary := &ScoreSummary{}
	blueScoreSummary := &ScoreSummary{}

	redScoreSummary.Score = 10
	blueScoreSummary.Score = 10
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))

	redScoreSummary.Score = 11
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.Score = 9
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	// Test playoff tiebreakers.
	redScoreSummary.Score = 10
	redScoreSummary.TowerPoints = 35
	blueScoreSummary.TowerPoints = 30
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.TowerPoints = 30
	blueScoreSummary.TowerPoints = 40
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, false))
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.TowerPoints = 30
	redScoreSummary.FuelPoints = 50
	blueScoreSummary.FuelPoints = 40
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.FuelPoints = 40
	blueScoreSummary.FuelPoints = 50
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.FuelPoints = 40
	redScoreSummary.NumOpponentMajorFouls = 2
	blueScoreSummary.NumOpponentMajorFouls = 0
	assert.Equal(t, RedWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	redScoreSummary.NumOpponentMajorFouls = 0
	blueScoreSummary.NumOpponentMajorFouls = 1
	assert.Equal(t, BlueWonMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))

	blueScoreSummary.NumOpponentMajorFouls = 0
	assert.Equal(t, TieMatch, DetermineMatchStatus(redScoreSummary, blueScoreSummary, true))
}
