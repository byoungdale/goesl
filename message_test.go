package goesl

import (
	"bufio"
	"strings"
	"testing"
)

// Build a bufio reader so we can mock esl's network reader
func reader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

var (
	// https://freeswitch.org/confluence/display/FREESWITCH/Event+List
	ShutdownMessage = `Content-Length: 4
Content-Type: text/event-plain
Event-Info: System Shutting Down
Event-Name: SHUTDOWN
Core-UUID: 596ab2fd-14c5-44b5-a02b-93ffb7cd5dd6
FreeSWITCH-Hostname: ********
FreeSWITCH-IPv4: ********
FreeSWITCH-IPv6: 127.0.0.1
Event-Date-Local: 2008-01-23 13:48:13
Event-Date-GMT: Wed,%2023%20Jan%202008%2018%3A48%3A13%20GMT
Event-Date-timestamp: 1201114093012795
Event-Calling-File: switch_core.c
Event-Calling-Function: switch_core_destroy
Event-Calling-Line-Number: 1046

done`
	EchoResponse = `Content-Type: api/response
Content-Length: 2

hi`

	HeartbeatMessage = `Event-Name: SOCKET_DATA
Content-Length: 965
Content-Type: text/event-json

{"Event-Name":"HEARTBEAT","Core-UUID":"09ae1707-6e50-4621-9bb6-d673aba7de08","FreeSWITCH-Hostname":"fs-server","FreeSWITCH-Switchname":"sip.example.com","FreeSWITCH-IPv4":"192.168.0.1","FreeSWITCH-IPv6":"::1","Event-Date-Local":"2023-10-03 14:13:36","Event-Date-GMT":"Tue, 03 Oct 2023 21:13:36 GMT","Event-Date-Timestamp":"1696367616134783","Event-Calling-File":"switch_core.c","Event-Calling-Function":"send_heartbeat","Event-Calling-Line-Number":"95","Event-Sequence":"1545554","Event-Info":"System Ready","Up-Time":"0 years, 19 days, 23 hours, 24 minutes, 59 seconds, 820 milliseconds, 134 microseconds","FreeSWITCH-Version":"1.10.10-release+git~20230812T150155Z~591f1eb749~64bit","Uptime-msec":"1725899820","Session-Count":"0","Max-Sessions":"2000","Session-Per-Sec":"30","Session-Per-Sec-Last":"0","Session-Per-Sec-Max":"10","Session-Per-Sec-FiveMin":"0","Session-Since-Startup":"4750","Session-Peak-Max":"11","Session-Peak-FiveMin":"0","Idle-CPU":"98.700000"}`
)

func TestNewMessage(t *testing.T) {
	buf := reader(HeartbeatMessage)
	fsMsg, err := NewMessage(buf, true)

	if err != nil {
		t.Error(err)
	}

	if fsMsg.Headers["FreeSWITCH-IPv4"] != "192.168.0.1" {
		t.Error("could not parse FreeSWITCH ip from event")
	}
}

func TestNewMessageMissingMime(t *testing.T) {
	heartbeatMimeless := strings.Replace(HeartbeatMessage, "Content-Type: text/event-json", "", 1)
	buf := reader(heartbeatMimeless)
	_, err := NewMessage(buf, true)

	if err == nil {
		t.Error("Expected error Parse EOF, got nothing")
		return
	}

	if !strings.Contains(err.Error(), "Parse EOF") {
		t.Error(err)
		return
	}
}

func TestNewMessageServerShutdown(t *testing.T) {
	buf := reader(ShutdownMessage)
	fsMsg, err := NewMessage(buf, true)

	if err != nil {
		t.Error(err)
	}

	if fsMsg.Headers["Content-Type"] != "text/event-plain" {
		t.Error("could not parse FreeSWITCH event")
	}

	if fsMsg.Body == nil {
		t.Error("Body is empty")
	}
}

func TestDump(t *testing.T) {
	buf := reader(ShutdownMessage)
	fsMsg, err := NewMessage(buf, true)

	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(fsMsg.Dump(), "Event-Info: System Shutting Down") {
		t.Error("freeswitch message dump failed")
	}
}

func TestMessageParse(t *testing.T) {
	buf := reader(EchoResponse)
	fsMsg, err := NewMessage(buf, true)

	if err != nil {
		t.Error(err)
	}

	body := string(fsMsg.Body)

	if body != "hi" {
		t.Error("parsing freeswitch response failed")
	}
}

func TestString(t *testing.T) {}

func TestGetCallUUID(t *testing.T) {}

func TestGetHeader(t *testing.T) {}

func TestParse(t *testing.T) {}
