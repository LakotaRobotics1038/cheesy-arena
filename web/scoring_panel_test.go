// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestScoringPanel(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/panels/scoring/invalidposition")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid position")
	recorder = web.getHttpResponse("/panels/scoring/red_near")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/red_far")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue_near")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue_far")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Scoring Panel - Untitled Event - Cheesy Arena")
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)
	redConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red_near/websocket", nil)
	assert.Nil(t, err)
	defer redConn.Close()
	redWs := websocket.NewTestWebsocket(redConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumPanels("blue_near"))
	blueConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blue_near/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := websocket.NewTestWebsocket(blueConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red_near"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("blue_near"))

	// Should get a few status updates right after connection.
	readWebsocketType(t, redWs, "resetLocalState")
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "resetLocalState")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "matchTime")
	readWebsocketType(t, blueWs, "realtimeScore")

	// Send some autonomous fuel scoring commands
	fuelData := struct {
		IsAuto bool
	}{}
	web.arena.MatchState = field.AutoPeriod
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel)
	assert.Equal(t, 0, web.arena.BlueRealtimeScore.CurrentScore.Hub.AutoFuel)
	fuelData.IsAuto = true
	redWs.Write("addFuel", fuelData)
	redWs.Write("addFuel", fuelData)
	redWs.Write("addFuel", fuelData)
	blueWs.Write("addFuel", fuelData)
	blueWs.Write("addFuel", fuelData)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, 3, web.arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel)
	assert.Equal(t, 2, web.arena.BlueRealtimeScore.CurrentScore.Hub.AutoFuel)

	// Test removing fuel
	redWs.Write("removeFuel", fuelData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, 2, web.arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel)

	// Test teleop fuel
	web.arena.MatchState = field.TeleopPeriod
	assert.Equal(t, 0, web.arena.RedRealtimeScore.CurrentScore.Hub.TeleopFuel)
	fuelData.IsAuto = false
	redWs.Write("addFuel", fuelData)
	redWs.Write("addFuel", fuelData)
	redWs.Write("addFuel", fuelData)
	redWs.Write("addFuel", fuelData)
	redWs.Write("addFuel", fuelData)
	for i := 0; i < 5; i++ {
		readWebsocketType(t, redWs, "realtimeScore")
		readWebsocketType(t, blueWs, "realtimeScore")
	}
	assert.Equal(t, 5, web.arena.RedRealtimeScore.CurrentScore.Hub.TeleopFuel)

	// Send some tower level scoring commands
	towerData := struct {
		Level  int
		IsAuto bool
	}{}
	assert.Equal(t, game.TowerLevelNone, web.arena.RedRealtimeScore.CurrentScore.Hub.AutoTowerLevel)
	assert.Equal(t, game.TowerLevelNone, web.arena.BlueRealtimeScore.CurrentScore.Hub.AutoTowerLevel)
	web.arena.MatchState = field.AutoPeriod
	towerData.Level = 1
	towerData.IsAuto = true
	redWs.Write("setTowerLevel", towerData)
	towerData.Level = 2
	blueWs.Write("setTowerLevel", towerData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, game.TowerLevel1, web.arena.RedRealtimeScore.CurrentScore.Hub.AutoTowerLevel)
	assert.Equal(t, game.TowerLevel2, web.arena.BlueRealtimeScore.CurrentScore.Hub.AutoTowerLevel)

	// Test teleop tower levels
	web.arena.MatchState = field.TeleopPeriod
	assert.Equal(t, game.TowerLevelNone, web.arena.RedRealtimeScore.CurrentScore.Hub.TeleopTowerLevel)
	towerData.Level = 3
	towerData.IsAuto = false
	redWs.Write("setTowerLevel", towerData)
	towerData.Level = 2
	blueWs.Write("setTowerLevel", towerData)
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, game.TowerLevel3, web.arena.RedRealtimeScore.CurrentScore.Hub.TeleopTowerLevel)
	assert.Equal(t, game.TowerLevel2, web.arena.BlueRealtimeScore.CurrentScore.Hub.TeleopTowerLevel)

	// Test that some invalid commands do nothing and don't result in score change notifications.
	redWs.Write("invalid", nil)
	towerData.Level = 10 // Invalid tower level
	redWs.Write("setTowerLevel", towerData)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))
	web.arena.MatchState = field.PostMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // Allow some time for the commands to be processed.
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))

	// Load another match to reset the results.
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, field.NewRealtimeScore(), web.arena.BlueRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red_near"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue_near"))
}
