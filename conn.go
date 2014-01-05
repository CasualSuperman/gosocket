package gosocket

import (
	"encoding/json"
	"io"
	"math/rand"
	"sync"

	ws "code.google.com/p/go.net/websocket"
)

var emptyHandlerMap = map[string][]Handler{}

// Conn is a connection between a server and a client.
type Conn struct {
	conn          *ws.Conn
	handlers      map[string][]Handler
	conversations map[int]chan message
	messageID     int
	open          bool
	lock          sync.Mutex
}

func (c *Conn) handleMessages(serverHandlers map[string][]Handler) error {
	if serverHandlers == nil {
		serverHandlers = emptyHandlerMap
	}

	for {
		msg, err := c.msg()

		if err != nil {
			if err == io.EOF {
				c.close()
			}
			return err
		}


		if msg.IsResp {
			if ch, ok := c.conversations[msg.ID]; ok {
				ch <- msg
			}
		} else {
			handlers := serverHandlers[msg.Path]

			for _, handler := range handlers {
				go handler(msg)
			}

			handlers = c.handlers[msg.Path]

			for _, handler := range handlers {
				go handler(msg)
			}
		}
	}
}

// Send a data structure on the given path.  It will be JSON-encoded.
func (c *Conn) Send(path string, data interface{}) (Msg, error) {
	msg, err := json.Marshal(data)

	if err != nil {
		return message{conn: c}, err
	}

	c.messageID++

	m := message{
		path,
		string(msg),
		c.messageID,
		false,
		c,
	}

	return m, c.sendMsg(m)
}

func (c *Conn) sendMsg(msg message) error {
	return ws.JSON.Send(c.conn, msg)
}

func (c *Conn) msg() (message, error) {
	var m message
	err := ws.JSON.Receive(c.conn, &m)
	m.conn = c
	return m, err
}

// Close a connection.
func (c *Conn) Close() error {
	c.close()
	return c.conn.Close()
}

func (c *Conn) close() {
	c.lock.Lock()
	c.open = false
	for _, ch := range c.conversations {
		close(ch)
	}
	c.lock.Unlock()
}

// Closed indicates if a connection has been closed.
func (c *Conn) Closed() bool {
	return !c.open
}

// Open a connection to the given location.
func Open(location string) (*Conn, error) {
	c, err := ws.Dial("ws://"+location, "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		c,
		make(map[string][]Handler),
		make(map[int]chan message),
		rand.Int(),
		true,
		mutex(),
	}
	go conn.handleMessages(nil)
	return conn, nil
}

// Handle allows adding extra handlers to individual connections.
func (c *Conn) Handle(path string, h Handler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	handlers := c.handlers[path]
	handlers = append(handlers, h)
	c.handlers[path] = handlers
}
