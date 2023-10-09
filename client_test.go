package goesl

import (
	"net"
	"strconv"
	"sync"
	"testing"
)

// Only testing the auth method because the rest is only a TCP client connection.
// <--server-- Content-Type: auth/request
// --client--> auth ClueCon
// <--server-- Content-Type: command/reply\r\nReply-Text: +OK accepted
func TestClientAuthenticate(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	client := Client{
		SocketConnection: SocketConnection{
			Conn: clientConn,
			mtx:  &sync.RWMutex{},
			err:  make(chan error),
			m:    make(chan *Message),
		},
		Proto:   "tcp", // Let me know if you ever need this open up lol
		Addr:    net.JoinHostPort("localhost", strconv.Itoa(int(8022))),
		Passwd:  "ClueCon",
		Timeout: 10,
	}

	defer serverConn.Close()
	defer clientConn.Close()

	// Server
	go func() {
		for {
			buf := make([]byte, 2048)
			// Check if this is the first data received from the client
			authResponse := "Content-Type: auth/request\r\n\r\n"
			serverConn.Write([]byte(authResponse))

			_, err := serverConn.Read(buf)
			if err != nil {
				t.Logf("Server: Error reading from client: '%v'", err)
				return
			}

			// Send +OK back
			response := "Content-Type: command/reply\r\nReply-Text: +OK accepted\r\n\r\n"
			serverConn.Write([]byte(response))
		}
	}()

	if err := client.Authenticate(); err != nil {
		t.Logf("Got error authenticating client: '%v'", err)
		t.Fail()
	}
}
