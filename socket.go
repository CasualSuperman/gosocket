package gosocket

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	ws "code.google.com/p/go.net/websocket"
)

func mutex() (s sync.Mutex) {
	return
}

type eventType byte

// Constants indicating event types that can be overridden within a server.
const (
	Connect eventType = iota
	Disconnect
)

// Handler is a handler for a given path within a server or connection.
type Handler func(Msg)

// Server is a gosocket server that can be plugged into an http server in the standard library.
type Server struct {
	handlers   map[string][]Handler
	connect    func(*Conn)
	disconnect func(*Conn)
	wsServer   ws.Server
	errorFunc  func(error)
}

// NewServer returns a new server with the default no-op handlers.
func NewServer() *Server {
	s := &Server{handlers: make(map[string][]Handler)}
	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))

	handleConn := func(wsConn *ws.Conn) {
		c := &Conn{
			wsConn,
			make(map[string][]Handler),
			make(map[int]chan message),
			randSrc.Int(),
			true,
			mutex(),
		}

		if s.connect != nil {
			s.connect(c)
		}

		for {
			msg, err := c.msg()

			if err == nil {
				if msg.IsResp {
					if ch, ok := c.conversations[msg.ID]; ok {
						ch <- msg
					}
				} else {
					handlers := s.handlers[msg.Path]
					for _, handler := range handlers {
						go handler(msg)
					}
					handlers = c.handlers[msg.Path]
					for _, handler := range handlers {
						go handler(msg)
					}
				}

			} else if err == io.EOF {
				c.open = false
				for _, ch := range c.conversations {
					close(ch)
				}
				if s.disconnect != nil {
					s.disconnect(c)
				}
				break
			} else if s.errorFunc != nil {
				s.errorFunc(err)
			}
		}
	}

	s.wsServer.Handler = handleConn

	return s
}

// Handle registers a handler for the given path.  More than one handler can be present on a given path, they will be called in parallel.
func (s *Server) Handle(path string, h Handler) {
	handlers := s.handlers[path]
	handlers = append(handlers, h)
	s.handlers[path] = handlers
}

// On allows a server to handle connection events. Each server can only have one handler for each event type.
func (s *Server) On(e eventType, f func(*Conn)) {
	switch e {
	case Connect:
		s.connect = f
	case Disconnect:
		s.disconnect = f
	}
}

// Errored is called when a server has a problem decoding a message.  This function is mainly useful for debugging.
func (s *Server) Errored(f func(error)) {
	s.errorFunc = f
}

// ServeHTTP allows a server to be added to an http.Server.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, "/gs.js") {
		w.Write([]byte(js))
	} else if strings.HasSuffix(req.URL.Path, "/gs.min.js") {
		w.Write([]byte(jsMin))
	} else {
		s.wsServer.ServeHTTP(w, req)
	}
}
