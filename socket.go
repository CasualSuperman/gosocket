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

const (
	Connect eventType = iota
	Disconnect
)

type Handler func(Msg)

type Server struct {
	lock       sync.Mutex
	handlers   map[string][]Handler
	connect    func(*Conn)
	disconnect func(*Conn)
	wsServer   ws.Server
	errorFunc  func(error)
}

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

func (s *Server) Handle(path string, h Handler) {
	s.lock.Lock()
	defer s.lock.Unlock()

	handlers := s.handlers[path]
	handlers = append(handlers, h)
	s.handlers[path] = handlers
}

func (s *Server) On(e eventType, f func(*Conn)) {
	switch e {
	case Connect:
		s.connect = f
	case Disconnect:
		s.disconnect = f
	}
}

func (s *Server) Errored(f func(error)) {
	s.errorFunc = f
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, "/gs.js") {
		w.Write([]byte(js))
	} else {
		s.wsServer.ServeHTTP(w, req)
	}
}
