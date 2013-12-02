package gosocket

import (
	"encoding/json"

	ws "code.google.com/p/go.net/websocket"
)

type conn struct {
	conn          *ws.Conn
	handlers      map[string][]Handler
	conversations map[int]chan message
	messageID     int
}

func (c *conn) Send(path string, data interface{}) error {
	msg, err := json.Marshal(data)

	if err != nil {
		return err
	}

	c.messageID++

	return c.sendMsg(message{
		path,
		string(msg),
		c.messageID,
		false,
		c,
	})
}

func (c *conn) sendMsg(msg message) error {
	return ws.JSON.Send(c.conn, msg)
}

func (c *conn) msg() (message, error) {
	m := message{conn: c}
	err := ws.JSON.Receive(c.conn, &m)
	return m, err
}
