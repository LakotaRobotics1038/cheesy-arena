// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package network

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigureSCC(t *testing.T) {
	username := "username"
	password := "password"
	upCommands := []string{
		"up_line1",
		"up line 2",
		"up-line/3",
		"up line 4",
		"up line 5",
		"exit",
	}
	downCommands := []string{
		"down_line1",
		"down line 2",
		"down-line/3",
		"down line 4",
		"down line 5",
		"exit",
	}
	scc := NewSCCSwitch("127.0.0.1", username, password, upCommands, downCommands)
	scc.port = 9150
	scc.connectTimeoutDuration = 10 * time.Millisecond
	scc.configTimeoutDuration = 15 * time.Millisecond

	var receivedUpCommands, receivedDownCommands []string

	// Set the switch to the down state
	mockTelnetSwitch(t, scc.port, username, password, &receivedDownCommands)
	assert.Nil(t, scc.SetTeamEthernetEnabled(false))
	assert.Equal(t, downCommands, receivedDownCommands)
	assert.Equal(t, "DISABLED", scc.Status)

	// Set the switch to the up state
	scc.port += 1
	mockTelnetSwitch(t, scc.port, username, password, &receivedUpCommands)
	assert.Nil(t, scc.SetTeamEthernetEnabled(true))
	assert.Equal(t, upCommands, receivedUpCommands)
	assert.Equal(t, "ACTIVE", scc.Status)
}
func mockTelnetSwitch(t *testing.T, port int, username, password string, commands *[]string) {
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		assert.Nil(t, err)
		defer listener.Close()
		conn, err := listener.Accept()
		assert.Nil(t, err)
		defer conn.Close()

		// Read all data sent by the client until a read timeout occurs.
		var receivedData bytes.Buffer
		buf := make([]byte, 1024)
		// Allow a short window for the client to send data.
		_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				receivedData.Write(buf[:n])
				// Extend deadline to keep reading as long as data keeps coming.
				_ = conn.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
			}
			if err != nil {
				netErr, ok := err.(net.Error)
				if ok && netErr.Timeout() {
					break
				}
				break
			}
		}

		*commands = strings.Split(receivedData.String(), "\n")
		if len(*commands) > 0 && (*commands)[len(*commands)-1] == "" {
			*commands = (*commands)[:len(*commands)-1] // Remove trailing newline
		}
	}()
	time.Sleep(50 * time.Millisecond) // Give it some time to open the socket.
}
