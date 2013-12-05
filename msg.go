package gosocket

import (
	"encoding/json"
	"time"
)

// Msg is a single message sent or received in a gosocket.
type Msg interface {
	// Receive allows a handler to retreive the sent data.
	Receive(interface{}) error

	// Respond allows a response to a message within the same thread of communication.
	Respond(interface{}) error

	// Response waits for a response to a message.
	Response() (Msg, error)
	// TimedResponse waits for a response to a message, but can time out.
	TimedResponse(time.Duration) (Msg, error)
}

type msgError string

var closeErr = msgError("connection closed")
var timeoutErr = msgError("operation timed out")

func (err msgError) Error() string {
	return string(err)
}

type message struct {
	Path   string
	Msg    string
	ID     int
	IsResp bool
	conn   *Conn
}

func (m message) Receive(v interface{}) error {
	return json.Unmarshal([]byte(m.Msg), v)
}

func (m message) Respond(data interface{}) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}

	m.Msg = string(msg)
	m.IsResp = true
	return m.conn.sendMsg(m)
}

func (m message) TimedResponse(timeout time.Duration) (Msg, error) {
	thread, ok := m.conn.conversations[m.ID]
	if !ok {
		thread = make(chan message)
		m.conn.conversations[m.ID] = thread
	}

	select {
	case <-time.After(timeout):
		return m, timeoutErr

	case msg, ok := <-thread:
		if ok {
			return msg, nil
		}
		return msg, closeErr
	}
}

func (m message) Response() (Msg, error) {
	thread, ok := m.conn.conversations[m.ID]

	if !ok {
		thread = make(chan message)
		m.conn.conversations[m.ID] = thread
	}

	msg, ok := <-thread

	if ok {
		return msg, nil
	}
	return msg, closeErr
}
