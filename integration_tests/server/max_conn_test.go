// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package server

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	commands "github.com/dicedb/dice/integration_tests/commands/resp"
)

func getConnection(port int) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TestMaxConnection(t *testing.T) {
	var wg sync.WaitGroup
	var maxConnTestOptions = commands.TestServerOptions{
		Port:       8741,
		MaxClients: 50,
	}
	commands.RunTestServer(&wg, maxConnTestOptions)

	time.Sleep(2 * time.Second)

	var maxConnLimit = maxConnTestOptions.MaxClients + 2
	connections := make([]net.Conn, maxConnLimit)
	defer func() {
		// Ensure all connections are closed at the end of the test
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	for i := int32(0); i < maxConnLimit; i++ {
		conn, err := getConnection(maxConnTestOptions.Port)
		if err == nil {
			connections[i] = conn
		} else {
			t.Fatalf("unexpected error while getting connection %d: %v", i, err)
		}
	}
	assert.Equal(t, maxConnLimit, int32(len(connections)), "should have reached the max connection limit")

	result := commands.FireCommand(connections[0], "ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	} else {
		slog.Info("Closed server for max_conn_test")
	}
	wg.Wait()
}
