// Copyright 2025 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring an SCC Switch via Telnet.

package network

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	sccSwitchConnectTimeoutSec = 5
	sccSwitchConfigTimeoutSec  = 5
	sccSwitchTelnetPort        = 23
)

type SCCSwitch struct {
	address                string
	port                   int
	username               string
	password               string
	mutex                  sync.Mutex
	connectTimeoutDuration time.Duration
	configTimeoutDuration  time.Duration
	upCommands             []string
	downCommands           []string
	Status                 string
}

func NewSCCSwitch(address, username, password string, upCommands, downCommands []string) *SCCSwitch {
	return &SCCSwitch{
		address:                address,
		port:                   sccSwitchTelnetPort,
		username:               username,
		password:               password,
		connectTimeoutDuration: sccSwitchConnectTimeoutSec * time.Second,
		configTimeoutDuration:  sccSwitchConfigTimeoutSec * time.Second,
		upCommands:             upCommands,
		downCommands:           downCommands,
		Status:                 "UNKNOWN",
	}
}

func (scc *SCCSwitch) SetTeamEthernetEnabled(enabled bool) error {
	scc.mutex.Lock()
	defer scc.mutex.Unlock()

	// If no address is configured, treat the SCC as disabled and skip configuration.
	if scc.address == "" {
		scc.Status = "DISABLED"
		return nil
	}

	scc.Status = "CONFIGURING"

	commandSequence := scc.downCommands
	if enabled {
		commandSequence = scc.upCommands
	}

	_, err := scc.runCommandSequence(commandSequence)
	if err != nil {
		scc.Status = "ERROR"
		return fmt.Errorf("failed to set team ethernet state: %w", err)
	}

	if enabled {
		scc.Status = "ACTIVE"
	} else {
		scc.Status = "DISABLED"
	}

	return nil
}

// Logs into the switch via Telnet and runs the given commands in sequence.
// Returns the output of the commands or an error if the operation fails.
func (scc *SCCSwitch) runCommandSequence(commands []string) (string, error) {
	// Open a Telnet (TCP) connection to the switch with a timeout.
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", scc.address, scc.port), scc.connectTimeoutDuration)
	if err != nil {
		return "", fmt.Errorf("failed to connect to switch: %w", err)
	}
	defer conn.Close()

	// Send the provided commands to the switch (no auth sequence for plain Telnet mock).
	writer := bufio.NewWriter(conn)
	for _, command := range commands {
		if _, err := writer.WriteString(command + "\n"); err != nil {
			return "", fmt.Errorf("failed to write command to switch: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("failed to flush commands to switch: %w", err)
	}

	// Close the write side so the server sees EOF and can finish processing.
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.CloseWrite()
	}

	// Give the device up to the config timeout to respond, then read whatever is available.
	deadline := time.Now().Add(scc.configTimeoutDuration)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return "", fmt.Errorf("failed to set read deadline: %w", err)
	}

	var reader bytes.Buffer
	_, err = reader.ReadFrom(conn)
	if err != nil {
		// If the read timed out but returned some data, return it; otherwise surface the error.
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			if reader.Len() > 0 {
				return reader.String(), nil
			}
			return "", fmt.Errorf("timed out waiting for command sequence to complete")
		}
		// If EOF or other error but we have data, return it.
		if err == io.EOF && reader.Len() > 0 {
			return reader.String(), nil
		}
		return "", err
	}

	return reader.String(), nil
}
