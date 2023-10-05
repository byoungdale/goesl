package goesl

import (
	"net"
	"testing"
)

func TestStartAndStop(t *testing.T) {
	// Create a new OutboundServer for testing
	server, err := NewOutboundServer("localhost:8021")
	if err != nil {
		t.Fatalf("Error creating OutboundServer: %v", err)
	}

	// Use a channel to signal when the server has started
	serverStarted := make(chan struct{})

	// Start the server
	go func() {
		defer close(serverStarted)
		if err := server.Start(); err != nil {
			t.Errorf("Error starting OutboundServer: %v", err)
		}
	}()

	conn, err := net.Dial("tcp", "localhost:8021")
	if err != nil {
		t.Errorf("Error making test connection to OutboundServer: %v", err)
		return
	}
	defer conn.Close()

	// Stop the server
	server.Stop()
}

func TestInvalidAddr(t *testing.T) {
	_, err := NewOutboundServer("")
	if err == nil {
		t.Error("Expected error for invalid address, but got nil")
	}
}
