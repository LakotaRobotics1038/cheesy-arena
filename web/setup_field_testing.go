// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for testing the field sounds, LEDs, and PLC.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
)

// Shows the Field Testing page.
func (web *Web) fieldTestingGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_field_testing.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	plc := web.arena.MainPlc
	data := struct {
		*model.EventSettings
		MatchSounds          []*game.MatchSound
		InputNames           []string
		RegisterNames        []string
		CoilNames            []string
		RedHubRegisterNames  []string
		RedHubCoilNames      []string
		BlueHubRegisterNames []string
		BlueHubCoilNames     []string
	}{web.arena.EventSettings, game.MatchSounds, plc.GetInputNames(), plc.GetRegisterNames(), plc.GetCoilNames(),
		web.arena.RedHubPlc.GetRegisterNames(), web.arena.RedHubPlc.GetCoilNames(),
		web.arena.BlueHubPlc.GetRegisterNames(), web.arena.BlueHubPlc.GetCoilNames()}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for sending realtime updates to the Field Testing page.
func (web *Web) fieldTestingWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MainPlc.IoChangeNotifier())
	go ws.HandleNotifiers(web.arena.RedHubPlc.IoChangeNotifier())
	go ws.HandleNotifiers(web.arena.BlueHubPlc.IoChangeNotifier())

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "playSound":
			sound, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			web.arena.PlaySoundNotifier.NotifyWithMessage(sound)
		case "setHubLightColor":
			params, ok := data.(map[string]interface{})
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			alliance, ok := params["alliance"].(string)
			if !ok {
				ws.WriteError("Invalid alliance parameter.")
				continue
			}
			red, ok := params["red"].(float64)
			if !ok {
				ws.WriteError("Invalid red parameter.")
				continue
			}
			green, ok := params["green"].(float64)
			if !ok {
				ws.WriteError("Invalid green parameter.")
				continue
			}
			blue, ok := params["blue"].(float64)
			if !ok {
				ws.WriteError("Invalid blue parameter.")
				continue
			}
			if alliance == "red" {
				web.arena.RedHubPlc.SetHubLightColor(uint16(red), uint16(green), uint16(blue))
			} else if alliance == "blue" {
				web.arena.BlueHubPlc.SetHubLightColor(uint16(red), uint16(green), uint16(blue))
			} else {
				ws.WriteError("Invalid alliance value. Must be 'red' or 'blue'.")
				continue
			}
		case "setHubActive":
			params, ok := data.(map[string]interface{})
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			alliance, ok := params["alliance"].(string)
			if !ok {
				ws.WriteError("Invalid alliance parameter.")
				continue
			}
			active, ok := params["active"].(bool)
			if !ok {
				ws.WriteError("Invalid active parameter.")
				continue
			}
			if alliance == "red" {
				web.arena.RedHubPlc.SetHubActive(active)
			} else if alliance == "blue" {
				web.arena.BlueHubPlc.SetHubActive(active)
			} else {
				ws.WriteError("Invalid alliance value. Must be 'red' or 'blue'.")
				continue
			}
		case "resetBallCount":
			alliance, ok := data.(string)
			if !ok {
				ws.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			if alliance == "red" {
				web.arena.RedHubPlc.SetHubCount(0)
			} else if alliance == "blue" {
				web.arena.BlueHubPlc.SetHubCount(0)
			} else {
				ws.WriteError("Invalid alliance value. Must be 'red' or 'blue'.")
				continue
			}
		default:
			ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}
	}
}
