// Copyright 2015 Nevio Vesic
// Please check out LICENSE file for more information about what you CAN and what you CANNOT do!
// Basically in short this is a free software for you to do whatever you want to do BUT copyright must be included!
// I didn't write all of this code so you could say it's yours.
// MIT License

package goesl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Message - Freeswitch Message that is received by GoESL. Message struct is here to help with parsing message
// and dumping its contents. In addition to that it's here to make sure received message is in fact message we wish/can support
type Message struct {
	Headers map[string]string
	Body    []byte

	r  *bufio.Reader
	tr *textproto.Reader
}

// String - Will return message representation as string
func (m *Message) String() string {
	return fmt.Sprintf("%v body=%s", m.Headers, m.Body)
}

// GetCallUUID - Will return Caller-Unique-ID
func (m *Message) GetCallUUID() string {
	return m.GetHeader("Caller-Unique-ID")
}

// GetHeader - Will return message header value, or "" if the key is not set.
func (m *Message) GetHeader(key string) string {
	return m.Headers[key]
}

// Parse - Will parse out message received from Freeswitch and basically build it accordingly for later use.
// However, in case of any issues func will return error.
func (m *Message) Parse() error {

	cmr, err := m.tr.ReadMIMEHeader()

	if err != nil && err.Error() != "EOF" {
		Error(ECouldNotReadMIMEHeaders, err)
		return err
	}

	if cmr.Get("Content-Type") == "" {
		Debug("Not accepting message because of empty content type. Just whatever with it ...")
		return fmt.Errorf("Parse EOF")
	}

	// Will handle content length by checking if appropriate length is here and if it is then
	// we are going to read it into body
	if lv := cmr.Get("Content-Length"); lv != "" {
		l, err := strconv.Atoi(lv)

		if err != nil {
			Error(EInvalidContentLength, err)
			return err
		}

		m.Body = make([]byte, l)

		// TODO
		// Add context here for timeout handling
		// If bad Content-Length is passed
		// this will hang forever
		if _, err := io.ReadFull(m.r, m.Body); err != nil {
			Error(ECouldNotReadyBody, err)
			return err
		}
	}

	msgType := cmr.Get("Content-Type")

	Debug("Got message content (type: %s). Searching if we can handle it ...", msgType)

	if !StringInSlice(msgType, AvailableMessageTypes) {
		return fmt.Errorf(EUnsupportedMessageType, msgType, AvailableMessageTypes)
	}

	// Assing message headers IF message is not type of event-json
	if msgType != "text/event-json" {
		for k, v := range cmr {

			m.Headers[k] = v[0]

			// Will attempt to decode if % is discovered within the string itself
			if strings.Contains(v[0], "%") {
				m.Headers[k], err = url.QueryUnescape(v[0])

				if err != nil {
					Error(ECouldNotDecode, err)
					continue
				}
			}
		}
	}

	switch msgType {
	case "text/disconnect-notice":
		for k, v := range cmr {
			Debug("Message (header: %s) -> (value: %v)", k, v)
		}
	case "command/reply":
		reply := cmr.Get("Reply-Text")

		if strings.Contains(reply, "-ERR") {
			return fmt.Errorf(EUnsuccessfulReply, reply[5:])
		}
	case "api/response":
		if strings.Contains(string(m.Body), "-ERR") {
			return fmt.Errorf(EUnsuccessfulReply, string(m.Body)[5:])
		}
	case "text/event-json":
		// OK, what is missing here is a way to interpret other JSON types - it expects string only (understandably
		// because the FS events are generally "string: string") - extract into empty interface and migrate only strings.
		// i.e. Event CHANNEL_EXECUTE_COMPLETE - "variable_DP_MATCH":["a=rtpmap:101 telephone-event/8000","101"]
		var decoded map[string]interface{}

		if err := json.Unmarshal(m.Body, &decoded); err != nil {
			return err
		}

		// Copy back in:
		for k, v := range decoded {
			switch v := v.(type) {
			case string:
				m.Headers[k] = v
			default:
				//delete(m.Headers, k)
				Warn("Removed non-string property (%s)", k)
			}
		}

		if v := m.Headers["_body"]; v != "" {
			m.Body = []byte(v)
			delete(m.Headers, "_body")
		} else {
			m.Body = []byte("")
		}

	case "text/event-plain":
		Debug("Content-Type is 'text/event-plain'. Nothing more to do with m.Body")
	}

	return nil
}

// Dump - Will return message prepared to be dumped out. It's like prettify message for output
func (m *Message) Dump() (resp string) {
	var keys []string

	for k := range m.Headers {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		resp += fmt.Sprintf("%s: %s\r\n", k, m.Headers[k])
	}

	resp += fmt.Sprintf("BODY: %v\r\n", string(m.Body))

	return
}

// NewMessage - Will build and execute parsing against received freeswitch message.
// As return will give brand new Message{} for you to use it.
func NewMessage(r *bufio.Reader, autoParse bool) (*Message, error) {

	msg := Message{
		r:       r,
		tr:      textproto.NewReader(r),
		Headers: make(map[string]string),
	}

	if autoParse {
		if err := msg.Parse(); err != nil {
			return &msg, err
		}
	}

	return &msg, nil
}
