// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestScoringPanel(t *testing.T) {
	web := setupTestWeb(t)

	recorder := web.getHttpResponse("/panels/scoring/invalidposition")
	assert.Equal(t, 500, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid position")
	recorder = web.getHttpResponse("/panels/scoring/red")
	assert.Equal(t, 200, recorder.Code)
	recorder = web.getHttpResponse("/panels/scoring/blue")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Scoring Panel - Untitled Event - Cheesy Arena")
}

func TestScoringPanelWebsocket(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	_, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blorpy/websocket", nil)
	assert.NotNil(t, err)
	redConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/red/websocket", nil)
	assert.Nil(t, err)
	defer redConn.Close()
	redWs := websocket.NewTestWebsocket(redConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))
	blueConn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/panels/scoring/blue/websocket", nil)
	assert.Nil(t, err)
	defer blueConn.Close()
	blueWs := websocket.NewTestWebsocket(blueConn)
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("red"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumPanels("blue"))

	// Should get a few status updates right after connection.
	readWebsocketType(t, redWs, "resetLocalState")
	readWebsocketType(t, redWs, "autoSavedStatus")
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "matchTime")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "resetLocalState")
	readWebsocketType(t, blueWs, "autoSavedStatus")
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
	// Each command sends realtimeScore to both panels + autoSavedStatus to sender only
	// Red's 3 commands + Blue's 2 commands = 5 realtimeScore messages to each panel
	// Red gets 3 autoSavedStatus, Blue gets 2 autoSavedStatus
	// Due to async notifier, order is non-deterministic, so read all messages
	redMessages := readWebsocketMultiple(t, redWs, 8)   // 5 realtimeScore + 3 autoSavedStatus
	blueMessages := readWebsocketMultiple(t, blueWs, 7) // 5 realtimeScore + 2 autoSavedStatus
	assert.NotNil(t, redMessages["realtimeScore"])
	assert.NotNil(t, redMessages["autoSavedStatus"])
	assert.NotNil(t, blueMessages["realtimeScore"])
	assert.NotNil(t, blueMessages["autoSavedStatus"])
	assert.Equal(t, 3, web.arena.RedRealtimeScore.CurrentScore.Hub.AutoFuel)
	assert.Equal(t, 2, web.arena.BlueRealtimeScore.CurrentScore.Hub.AutoFuel)

	// Test removing fuel
	redWs.Write("removeFuel", fuelData)
	// Red gets realtimeScore + autoSavedStatus, Blue gets realtimeScore
	redMessages = readWebsocketMultiple(t, redWs, 2)
	blueMessages = readWebsocketMultiple(t, blueWs, 1)
	assert.NotNil(t, redMessages["realtimeScore"])
	assert.NotNil(t, redMessages["autoSavedStatus"])
	assert.NotNil(t, blueMessages["realtimeScore"])
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
	// Red sends 5 commands, each sends realtimeScore to both panels + autoSavedStatus to red
	redMessages = readWebsocketMultiple(t, redWs, 10)  // 5 realtimeScore + 5 autoSavedStatus
	blueMessages = readWebsocketMultiple(t, blueWs, 5) // 5 realtimeScore
	assert.NotNil(t, redMessages["realtimeScore"])
	assert.NotNil(t, redMessages["autoSavedStatus"])
	assert.NotNil(t, blueMessages["realtimeScore"])
	assert.Equal(t, 5, web.arena.RedRealtimeScore.CurrentScore.Hub.TeleopFuel)

	// Test that some invalid commands do nothing and don't result in score change notifications.
	redWs.Write("invalid", nil)

	// Test committing logic.
	redWs.Write("commitMatch", nil)
	readWebsocketType(t, redWs, "error")
	blueWs.Write("commitMatch", nil)
	readWebsocketType(t, blueWs, "error")
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
	web.arena.MatchState = field.PostMatch
	redWs.Write("commitMatch", nil)
	blueWs.Write("commitMatch", nil)
	time.Sleep(time.Millisecond * 10) // Allow some time for the commands to be processed.
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 1, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))

	// Load another match to reset the results.
	web.arena.ResetMatch()
	web.arena.LoadTestMatch()
	readWebsocketType(t, redWs, "matchLoad")
	readWebsocketType(t, redWs, "realtimeScore")
	readWebsocketType(t, blueWs, "matchLoad")
	readWebsocketType(t, blueWs, "realtimeScore")
	assert.Equal(t, field.NewRealtimeScore(), web.arena.RedRealtimeScore)
	assert.Equal(t, field.NewRealtimeScore(), web.arena.BlueRealtimeScore)
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("red"))
	assert.Equal(t, 0, web.arena.ScoringPanelRegistry.GetNumScoreCommitted("blue"))
}
