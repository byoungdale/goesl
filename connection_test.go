package goesl

import (
	"bufio"
	"bytes"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	c := &SocketConnection{
		mtx: &sync.RWMutex{},
	}

	_, err := c.Dial("127.0.0.1", "8021", time.Duration(10))
	if err == nil {
		t.Fatal("Expected non-nil error")
		t.Fail()
	}
}

func TestSend(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Test valid command
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := serverConn.Read(buf)
			if err != nil {
				t.Logf("Server: Error reading from client: '%v'", err)
				return
			}
			t.Logf("Server: Received %s from client\n", buf[:n])
		}
	}()

	err := c.Send("auth ClueCon")

	if err != nil {
		t.Logf("Got error sending request: '%v'", err)
		t.Fail()
	}

	err = c.Send("should error\r\n")

	if err == nil {
		t.Fatal("Expeceted non-nil err")
		t.Fail()
	}
}

func TestSendMany(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Test valid command
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := serverConn.Read(buf)
			if err != nil {
				t.Logf("Server: Error reading from client: '%v'", err)
				return
			}
			t.Logf("Server: Received %s from client\n", buf[:n])
		}
	}()

	validCmds := []string{"auth ClueCon", "log debug"}

	err := c.SendMany(validCmds)

	if err != nil {
		t.Logf("Got error sending request: '%v'", err)
		t.Fail()
	}

	withInvalidCmd := []string{"auth ClueCon", "log debug", "exploit\r\n"}

	err = c.SendMany(withInvalidCmd)

	if err == nil {
		t.Fatal("Expeceted non-nil err")
		t.Fail()
	}
}

func TestSendEvent(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Test valid command
	go func() {
		t.Log("Server: Received from client\n")
		for {
			buf := make([]byte, 2048)
			n, err := serverConn.Read(buf)
			if err != nil {
				t.Logf("Server: Error reading from client: '%v'", err)
				return
			}
			t.Logf("%s", buf[:n])
		}
	}()

	// Test sending with no headers
	emptyHeaders := []string{}
	err := c.SendEvent("SEND_INFO", emptyHeaders, "")

	if err == nil || err.Error() != "Must send at least one event header, detected `0` header" {
		t.Logf("Got error sending request: '%v'", err)
		t.Fail()
	}

	// Test sending with headers
	// CHANNEL_HANGUP event Event headers
	// reference: https://developer.signalwire.com/freeswitch/FreeSWITCH-Explained/Introduction/Event-System/Event-List_7143557/#16-channel_hangup-event
	headers := []string{
		"Hangup-Cause: NORMAL_CLEARING",
		"Channel-Read-Codec-Name: PCMU",
		"Channel-Read-Codec-Rate: 8000",
		"Channel-Write-Codec-Name: PCMU",
		"Channel-Write-Codec-Rate: 8000",
		"Caller-Username: jonas",
		"Caller-Dialplan: XML",
		"Caller-Caller-ID-Name: jonas",
		"Caller-Caller-ID-Number: jonas",
		"Caller-Network-Addr: 192.168.0.58",
		"Caller-Destination-Number: 541",
		"Caller-Unique-ID: 0dd4e4f7-36ed-a04d-a8f7-7aebb683af50",
		"Caller-Source: mod_sofia",
		"Caller-Context: default",
		"Caller-Channel-Name: sofia/192.168.0.58/jonas%40192.168.0.58%3A5060",
		"Caller-Screen-Bit: yes",
		"Caller-Privacy-Hide-Name: no",
		"Caller-Privacy-Hide-Number: no",
		"Originatee-Username: jonas",
		"Originatee-Dialplan: XML",
		"Originatee-Caller-ID-Name: jonas",
		"Originatee-Caller-ID-Number: jonas",
		"Originatee-Network-Addr: 192.168.0.58",
		"Originatee-Destination-Number: 192.168.0.58/arne%25192.168.0.58",
		"Originatee-Unique-ID: f66e8e31-c9fb-9b41-a9a2-a1586facb97f",
		"Originatee-Source: mod_sofia",
		"Originatee-Context: default",
		"Originatee-Channel-Name: sofia/192.168.0.58/arne",
		"Originatee-Screen-Bit: yes",
		"Originatee-Privacy-Hide-Name: no",
		"Originatee-Privacy-Hide-Number: no",
	}

	err = c.SendEvent("CHANNEL_HANGUP", headers, "")

	if err != nil {
		t.Logf("Got error sending request: '%v'", err)
		t.Fail()
	}

	// Test sending body
	// SEND_INFO
	// profile: external
	// content-type: text/plain
	// to-uri: sip:1@2.3.4.5
	// from-uri: sip:1@1.2.3.4
	// content-length: 15
	//
	// test
	headers = []string{
		"Hangup-Cause: NORMAL_CLEARING",
		"Channel-Read-Codec-Name: PCMU",
		"Channel-Read-Codec-Rate: 8000",
	}

	err = c.SendEvent("SEND_INFO", headers, "test")

	if err != nil {
		t.Logf("Got error sending request: '%v'", err)
		t.Fail()
	}
}

func TestSendMsg(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Server
	go func() {
		t.Log("Server: Received from client\n")
		for {
			buf := make([]byte, 2048)
			n, err := serverConn.Read(buf)
			if err != nil {
				t.Logf("Server: Error reading from client: '%v'", err)
				return
			}

			t.Logf("%s", buf[:n])
			// Simulate a response from the server to the client
			response := "Content-Type: command/reply\r\nReply-Text: +OK Job-UUID: c3b923ab-11c9-4063-bede-f6dedafb91ed\r\n\r\n"
			serverConn.Write([]byte(response))
		}
	}()

	// Client
	go func() {
		t.Log("Client: Received from server\n")
		for {
			buf := make([]byte, 2048)
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Logf("Client: Error reading from client: '%v'", err)
				c.err <- err
			}

			t.Logf("%s", buf[:n])

			// Create a *bytes.Buffer and write the byte data into it
			buffer := bytes.NewBuffer(buf[:n])

			// Create a *bufio.Reader that reads from the *bytes.Buffer
			reader := bufio.NewReader(buffer)

			m, err := newMessage(reader, true)
			if err != nil {
				t.Log("Problem parsing message")
				c.err <- err
			}
			c.m <- m
		}
	}()

	// sendmsg <uuid>
	// call-command: execute
	// execute-app-name: playback
	// execute-app-arg: /tmp/test.wav

	msg, err := c.SendMsg(map[string]string{
		"call-command":     "execute",
		"execute-app-name": "playback",
		"execute-app-arg":  "/tmp/test.wav",
	}, "c3b923ab-11c9-4063-bede-f6dedafb91ed", "")

	if err != nil {
		t.Logf("Got error sending request: '%v'", err)
		t.Fail()
	}

	if msg == nil {
		t.Log("Expected non-nil result in 'msg'")
		t.Fail()
	}
}

func TestReadMsg(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Server
	go func() {
		t.Log("Sending event from server")
		// Simulate a event from the server to the client
		event := "Content-Length: 907\r\nContent-Type: text/event-plain\r\n\r\nHangup-Cause: NORMAL_CLEARING\r\nChannel-Read-Codec-Name: PCMU\r\nChannel-Read-Codec-Rate: 8000\r\nChannel-Write-Codec-Name: PCMU\r\nChannel-Write-Codec-Rate: 8000\r\nCaller-Username: jonas\r\nCaller-Dialplan: XML\r\nCaller-Caller-ID-Name: jonas\r\nCaller-Caller-ID-Number: jonas\r\nCaller-Network-Addr: 192.168.0.58\r\nCaller-Destination-Number: 541\r\nCaller-Unique-ID: 0dd4e4f7-36ed-a04d-a8f7-7aebb683af50\r\nCaller-Source: mod_sofia\r\nCaller-Context: default\r\nCaller-Screen-Bit: yes\r\nCaller-Privacy-Hide-Name: no\r\nCaller-Privacy-Hide-Number: no\r\nOriginatee-Username: jonas\r\nOriginatee-Dialplan: XML\r\nOriginatee-Caller-ID-Name: jonas\r\nOriginatee-Caller-ID-Number: jonas\r\nOriginatee-Network-Addr: 192.168.0.58\r\nOriginatee-Unique-ID: f66e8e31-c9fb-9b41-a9a2-a1586facb97f\r\nOriginatee-Source: mod_sofia\r\nOriginatee-Context: default\r\nOriginatee-Screen-Bit: yes\r\nOriginatee-Privacy-Hide-Name: no\r\nOriginatee-Privacy-Hide-Number: no\r\n\r\n"
		serverConn.Write([]byte(event))
	}()

	// Client
	go func() {
		t.Log("Client: Received from server\n")
		for {
			buf := make([]byte, 2048)
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Logf("Client: Error reading from client: '%v'", err)
				c.err <- err
			}

			// Create a *bytes.Buffer and write the byte data into it
			buffer := bytes.NewBuffer(buf[:n])

			// Create a *bufio.Reader that reads from the *bytes.Buffer
			reader := bufio.NewReader(buffer)

			m, err := newMessage(reader, true)
			if err != nil {
				t.Log("Problem parsing message")
				c.err <- err
			}
			c.m <- m
		}
	}()

	t.Log("Going to read message")
	msg, err := c.ReadMsg()
	t.Log("client got message from server")

	if err != nil {
		t.Logf("Got error from ReadMsg: '%v'", err)
	}

	if msg == nil {
		t.Fatal("Expected msg to be non-nil")
	}
}

func TestExecute(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Server
	go func() {
		t.Log("Server: Received from client\n")
		for {
			buf := make([]byte, 2048)
			n, err := serverConn.Read(buf)
			if err != nil {
				t.Logf("Server: Error reading from client: '%v'", err)
				return
			}

			t.Logf("%s", buf[:n])
			// Simulate a response from the server to the client
			response := "Content-Type: command/reply\r\nReply-Text: +OK\r\n\r\n"
			serverConn.Write([]byte(response))
		}
	}()

	// Client
	go func() {
		t.Log("Client: Received from server\n")
		for {
			buf := make([]byte, 2048)
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Logf("Client: Error reading from client: '%v'", err)
				c.err <- err
			}

			t.Logf("%s", buf[:n])

			// Create a *bytes.Buffer and write the byte data into it
			buffer := bytes.NewBuffer(buf[:n])

			// Create a *bufio.Reader that reads from the *bytes.Buffer
			reader := bufio.NewReader(buffer)

			m, err := newMessage(reader, true)
			if err != nil {
				t.Log("Problem parsing message")
				c.err <- err
			}
			c.m <- m
		}
	}()

	// sendmsg
	// call-command: execute
	// execute-app-name: playback
	// execute-app-arg: /tmp/test.wav\n\n
	if _, err := c.Execute("playback", "/tmp/test.wav", true); err != nil {
		t.Fatalf("Got error while executing playback: %s", err)
	}
}

func TestExecuteUUID(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Server
	go func() {
		t.Log("Server: Received from client\n")
		for {
			buf := make([]byte, 2048)
			n, err := serverConn.Read(buf)
			if err != nil {
				t.Logf("Server: Error reading from client: '%v'", err)
				return
			}

			t.Logf("%s", buf[:n])
			// Simulate a response from the server to the client
			response := "Content-Type: command/reply\r\nReply-Text: +OK\r\n\r\n"
			serverConn.Write([]byte(response))
		}
	}()

	// Client
	go func() {
		t.Log("Client: Received from server\n")
		for {
			buf := make([]byte, 2048)
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Logf("Client: Error reading from client: '%v'", err)
				c.err <- err
			}

			t.Logf("%s", buf[:n])

			// Create a *bytes.Buffer and write the byte data into it
			buffer := bytes.NewBuffer(buf[:n])

			// Create a *bufio.Reader that reads from the *bytes.Buffer
			reader := bufio.NewReader(buffer)

			m, err := newMessage(reader, true)
			if err != nil {
				t.Log("Problem parsing message")
				c.err <- err
			}
			c.m <- m
		}
	}()

	// sendmsg <uuid>
	// call-command: execute
	// execute-app-name: playback
	// execute-app-arg: /tmp/test.wav
	// event-lock: true
	if _, err := c.ExecuteUUID("c3b923ab-11c9-4063-bede-f6dedafb91ed", "playback", "/tmp/test.wav", true); err != nil {
		t.Fatalf("Got error while executing playback: %s", err)
	}
}

// Going to be a pipe in test cases
// Probably need better testing here
func TestOriginatorAddr(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	addr := c.OriginatorAddr()
	if addr == nil {
		t.Fatal("couldn't get an address")
	}
}

func TestHandle(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	c := &SocketConnection{
		Conn: clientConn,
		mtx:  &sync.RWMutex{},
		err:  make(chan error),
		m:    make(chan *Message),
	}
	defer c.Close()
	defer serverConn.Close()
	defer clientConn.Close()

	// Server
	go func() {
		t.Log("Sending event from server")
		// Simulate a event from the server to the client
		event := "Content-Length: 907\r\nContent-Type: text/event-plain\r\n\r\nHangup-Cause: NORMAL_CLEARING\r\nChannel-Read-Codec-Name: PCMU\r\nChannel-Read-Codec-Rate: 8000\r\nChannel-Write-Codec-Name: PCMU\r\nChannel-Write-Codec-Rate: 8000\r\nCaller-Username: jonas\r\nCaller-Dialplan: XML\r\nCaller-Caller-ID-Name: jonas\r\nCaller-Caller-ID-Number: jonas\r\nCaller-Network-Addr: 192.168.0.58\r\nCaller-Destination-Number: 541\r\nCaller-Unique-ID: 0dd4e4f7-36ed-a04d-a8f7-7aebb683af50\r\nCaller-Source: mod_sofia\r\nCaller-Context: default\r\nCaller-Screen-Bit: yes\r\nCaller-Privacy-Hide-Name: no\r\nCaller-Privacy-Hide-Number: no\r\nOriginatee-Username: jonas\r\nOriginatee-Dialplan: XML\r\nOriginatee-Caller-ID-Name: jonas\r\nOriginatee-Caller-ID-Number: jonas\r\nOriginatee-Network-Addr: 192.168.0.58\r\nOriginatee-Unique-ID: f66e8e31-c9fb-9b41-a9a2-a1586facb97f\r\nOriginatee-Source: mod_sofia\r\nOriginatee-Context: default\r\nOriginatee-Screen-Bit: yes\r\nOriginatee-Privacy-Hide-Name: no\r\nOriginatee-Privacy-Hide-Number: no\r\n\r\n"
		serverConn.Write([]byte(event))
	}()

	// Client
	go func() {
		t.Log("Client: Received from server\n")
		for {
			buf := make([]byte, 2048)
			n, err := clientConn.Read(buf)
			if err != nil {
				t.Logf("Client: Error reading from client: '%v'", err)
				c.err <- err
			}

			// Create a *bytes.Buffer and write the byte data into it
			buffer := bytes.NewBuffer(buf[:n])

			// Create a *bufio.Reader that reads from the *bytes.Buffer
			reader := bufio.NewReader(buffer)

			m, err := newMessage(reader, true)
			if err != nil {
				t.Log("Problem parsing message")
				c.err <- err
			}
			c.m <- m
		}
	}()

	c.Handle()
}

func TestClose(t *testing.T) {
	t.FailNow()
}

func TestConnected(t *testing.T) {
	t.FailNow()
}

func TestReconnectIfNeeded(t *testing.T) {
	t.FailNow()
}
