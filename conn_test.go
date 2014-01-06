package gosocket

import (
	"net/http"
	"testing"
	"time"
	"sync"
)

var connServer *Server

func init() {
	connServer = startConnServer()
}

func startConnServer() *Server {
	s := NewServer()
	go func() {
		err := http.ListenAndServe(":6669", s)
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	return s
}

func connectConnServer() (*Conn, error) {
	return Open("localhost:6669")
}

func clearServer(s *Server) {
	for key := range s.handlers {
		delete(s.handlers, key)
	}
	s.On(Connect, nil)
	s.On(Disconnect, nil)
}

func clearConnServer() {
	clearServer(connServer)
}

func TestConnOpenFail(t *testing.T) {
	_, err := Open("localhost:8000")
	if err == nil {
		t.Error("connection should have failed")
	}
}

func TestConnOpen(t *testing.T) {
	defer clearConnServer()
	var wg sync.WaitGroup
	connected := false

	wg.Add(1)
	connServer.On(Connect, func(c *Conn) {
		connected = true
		wg.Done()
	})

	conn, err := connectConnServer()
	if err != nil {
		t.Error(err)
	}

	wg.Wait()
	if !connected {
		t.Error("failed to detect connection")
	}

	connServer.On(Connect, nil)
	conn.Close()
}

func TestConnSend(t *testing.T) {
	defer clearConnServer()
	done := make(chan bool)

	connServer.Handle("echo", func(m Msg) {
		var s interface{}
		err := m.Receive(&s)
		if err != nil {
			t.Fail()
		}
		done <- true
	})

	conn, _ := connectConnServer()
	conn.Send("echo", "Hello, world")

	<-done
	conn.Close()
}

func TestConnSendUnsendable(t *testing.T) {
	defer clearConnServer()
	conn, _ := connectConnServer()
	_, err := conn.Send("wherever", func(){})
	if err == nil {
		t.Error("can't encode a function to JSON")
	}
}

func TestConnClose(t *testing.T) {
	defer clearConnServer()
	conn, _ := connectConnServer()
	err := conn.Close()
	if err != nil {
		t.Error("closing failed: " + err.Error())
	}
}

func TestConnClosed(t *testing.T) {
	defer clearConnServer()
	conn, _ := connectConnServer()
	if conn.Closed() {
		t.Error("conn should not be closed")
	}
	conn.Close()
	if !conn.Closed() {
		t.Error("conn should be closed")
	}
}

func TestConnHandle(t *testing.T) {
	defer clearConnServer()
	connServer.On(Connect, func(c *Conn) {
		c.Send("hello", "world")
	})

	done := make(chan bool)
	conn, _ := connectConnServer()

	conn.Handle("hello", func(m Msg) {
		var s string
		m.Receive(&s)
		if s != "world" {
			t.Error("message incorrect")
		}
		done <- true
	})

	select {
	case <-done:
		conn.Close()
	case <-time.After(1000 * time.Millisecond):
		t.Error("message not received")
	}
	connServer.On(Connect, nil)
}

func TestConnResponse(t *testing.T) {
	defer clearConnServer()
	conn, _ := connectConnServer()
	connServer.Handle("echo", func(m Msg) {
		var data interface{}
		m.Receive(&data)
		m.Respond(data)
	})
	msg, _ := conn.Send("echo", "Hello, world!")
	msg, err := msg.Response()
	if err != nil {
		t.Error("response failed to arrive")
	}
	var s string
	msg.Receive(&s)
	if s != "Hello, world!" {
		t.Error("response didn't match")
	}
}
