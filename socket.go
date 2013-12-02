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

type Handler func(Msg)

type Server struct {
	lock      sync.Mutex
	handlers  map[string][]Handler
	wsServer  ws.Server
	closeFunc func(*Conn)
	errorFunc func(error)
}

func NewServer() *Server {
	s := &Server{handlers: make(map[string][]Handler)}
	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))

	handleConn := func(conn *ws.Conn) {
		c := &Conn{
			conn,
			make(map[string][]Handler),
			make(map[int]chan message),
			randSrc.Int(),
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
				}

			} else if err == io.EOF {
				for _, ch := range c.conversations {
					close(ch)
				}
				if s.closeFunc != nil {
					s.closeFunc(c)
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

func (s *Server) Closed(f func(*Conn)) {
	s.closeFunc = f
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
