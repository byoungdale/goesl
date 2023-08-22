package goesl

import (
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
	}

	err = c.Send("should error\r\n")

	if err == nil {
		t.Fatal("Expeceted non-nil err")
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
	}

	withInvalidCmd := []string{"auth ClueCon", "log debug", "exploit\r\n"}

	err = c.SendMany(withInvalidCmd)

	if err == nil {
		t.Fatal("Expeceted non-nil err")
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
	}
}

func TestExecute(t *testing.T) {
	t.FailNow()
}

func TestUUID(t *testing.T) {
	t.FailNow()
}

func TestMsg(t *testing.T) {
	t.FailNow()
}

func TestOriginatorAddr(t *testing.T) {
	t.FailNow()
}

func TestReadMessage(t *testing.T) {
	t.FailNow()
}

func TestHandle(t *testing.T) {
	t.FailNow()
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
