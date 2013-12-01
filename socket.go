package gosocket

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	ws "code.google.com/p/go.net/websocket"
)

type Server struct {
	lock      sync.Mutex
	handlers  map[string][]Handler
	wsServer  ws.Server
	message   serverMessage
	closeFunc func(Conn)
	errorFunc func(error)
}

type serverMessage struct {
	Path string
	Msg  string
}

func NewServer() *Server {
	s := &Server{handlers: make(map[string][]Handler)}

	handleConn := func(conn *ws.Conn) {
		c := Conn{conn}
		open := true

		for open {
			err := ws.JSON.Receive(conn, &s.message)

			if err == nil {
				handlers := s.handlers[s.message.Path]
				for _, handler := range handlers {
					go handler(c, Data(s.message.Msg))
				}
			} else if err == io.EOF {
				if s.closeFunc != nil {
					s.closeFunc(c)
				}
				open = false
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

func (s *Server) Closed(f func(Conn)) {
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

const js = `function GoSocket(url) {
	var conn = new WebSocket("ws://" + url);
	var paths = {};

	conn.onmessage = function(msg) {
		var msg = JSON.parse(msg.data);
		var handlers = paths[msg.Path];

		if (handlers) {
			var data = JSON.parse(msg.Msg);
			for (var i = 0; i < handlers.length; i++) {
				handlers[i](data);
			}
		}
	}

	this.Ready = function(func) {
		if (conn.readyState >= conn.OPEN) {
			func();
		} else {
			conn.onopen = func
		}
	}

	this.Send = function(path, msg) {
		conn.send(JSON.stringify({
			Path: path,
			Msg: JSON.stringify(msg),
		}));
	};

	this.On = function(path, func) {
		var handlers = paths[path];
		if (!handlers) {
			handlers = [];
		}
		handlers.push(func);
		paths[path] = handlers;
	};

	this.Close = function() {
		conn.close()
	};

	this._conn = conn;

	return this;
};`

type Conn struct {
	c *ws.Conn
}

func (c Conn) Send(path string, data interface{}) error {
	msg, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return ws.JSON.Send(c.c, serverMessage{
		path,
		string(msg),
	})
}

type Data []byte

func (d Data) Receive(v interface{}) error {
	return json.Unmarshal([]byte(d), v)
}

type Handler func(Conn, Data)
