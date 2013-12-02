package gosocket

import (
	"encoding/json"
	"time"
)

type Msg interface {
	Receive(interface{}) error

	Respond(interface{}) error

	Response() (Msg, error)
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
	return Data(m.Msg).Receive(v)
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
