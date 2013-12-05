package gosocket

import (
	"encoding/json"
	"math/rand"
	"sync"

	ws "code.google.com/p/go.net/websocket"
)

type Conn struct {
	conn          *ws.Conn
	handlers      map[string][]Handler
	conversations map[int]chan message
	messageID     int
	open          bool
	lock          sync.Mutex
}

func (c *Conn) Send(path string, data interface{}) error {
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

func (c *Conn) sendMsg(msg message) error {
	return ws.JSON.Send(c.conn, msg)
}

func (c *Conn) msg() (message, error) {
	m := message{conn: c}
	err := ws.JSON.Receive(c.conn, &m)
	return m, err
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) Closed() bool {
	return !c.open
}

func Open(location string) (*Conn, error) {
	c, err := ws.Dial("ws://"+location, "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	return &Conn{
		c,
		make(map[string][]Handler),
		make(map[int]chan message),
		rand.Int(),
		true,
		mutex(),
	}, nil
}

func (c *Conn) Handle(path string, h Handler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	handlers := c.handlers[path]
	handlers = append(handlers, h)
	c.handlers[path] = handlers
}
