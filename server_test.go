package goesl

import (
	"net"
	"os"
	"testing"
)

func TestStartAndStop(t *testing.T) {
	// Create a new OutboundServer for testing
	server, err := NewOutboundServer("127.0.0.1:8021")
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

	// Check if running in GitHub Actions environment
	if os.Getenv("CI") == "true" {
		// Use 'app' name in GitHub Actions environment
		conn, err := net.Dial("tcp", "app:8021")
		if err != nil {
			t.Errorf("Error making test connection to OutboundServer: %v", err)
			return
		}
		defer conn.Close()
	} else {
		conn, err := net.Dial("tcp", "127.0.0.1:8021")
		if err != nil {
			t.Errorf("Error making test connection to OutboundServer: %v", err)
			return
		}
		defer conn.Close()
	}

	// Stop the server
	server.Stop()
}

func TestInvalidAddr(t *testing.T) {
	_, err := NewOutboundServer("")
	if err == nil {
		t.Error("Expected error for invalid address, but got nil")
	}
}
