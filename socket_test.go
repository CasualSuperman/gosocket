package gosocket

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var socketServer *Server

func init() {
	socketServer = startSocketTestServer()
}

func startSocketTestServer() *Server {
	s := NewServer()
	go func() {
		err := http.ListenAndServe(":6668", s)
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	return s
}

func connectSocketTestServer() (*Conn, error) {
	return Open("localhost:6668")
}

func clearSocketServer() {
	clearServer(socketServer)
}

func TestServerBindFail(t *testing.T) {
	connected := make(chan error)
	go func() {
		connected <- http.ListenAndServe(":2p", NewServer())
	}()

	select {
	case err := <-connected:
		if err == nil {
			t.Error("bind returned but with no error")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("binded correctly to poorly formatted port")
	}
}

func TestServerConnect(t *testing.T) {
	defer clearSocketServer()
	done := make(chan bool)
	socketServer.On(Connect, func(c *Conn) {
		done <- true
	})

	connectSocketTestServer()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("connect handler never called")
	}
}

func TestServerDisconnect(t *testing.T) {
	defer clearSocketServer()
	done := make(chan bool)
	socketServer.On(Disconnect, func(c *Conn) {
		done <- true
	})

	c, _ := connectSocketTestServer()
	c.Close()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("connect handler never called")
	}
}

func TestServerDisconnectWithConversations(t *testing.T) {
	defer clearSocketServer()
	handlerCalled := make(chan bool)
	done := make(chan bool)

	socketServer.Handle("one", func(m Msg) {
		handlerCalled <- true
		_, err := m.Response()
		if err == nil {
			t.Error("response should fail")
		}
		done <- true
	})

	c, _ := connectSocketTestServer()
	c.Send("one", "test")
	<-handlerCalled
	c.Close()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("response should have been cancelled")
	}
}

func TestServerServeJS(t *testing.T) {
	page, err := http.Get("http://localhost:6668/gs.js")
	if err != nil {
		t.Error("could not find javascript on server")
		return
	}
	defer page.Body.Close()
	body, _ := ioutil.ReadAll(page.Body)
	if string(body) != js {
		t.Error("server javascript data doesn't match response")
	}
}

func TestServerServeJSMin(t *testing.T) {
	page, err := http.Get("http://localhost:6668/gs.min.js")
	if err != nil {
		t.Error("could not find javascript on server")
		return
	}
	defer page.Body.Close()
	body, _ := ioutil.ReadAll(page.Body)
	if string(body) != jsMin {
		t.Error("server javascript data doesn't match response")
	}
}

func TestServerHandle(t *testing.T) {
	defer clearSocketServer()
	done := make(chan bool)
	socketServer.Handle("one", func(m Msg) {
		done <- true
	})
	socketServer.Handle("two", func(m Msg) {
		done <- false
	})

	c, _ := connectSocketTestServer()
	c.Send("one", nil)
	c.Send("two", nil)
	resp := false
	for i := 0; i < 2; i++ {
		resp = resp != <-done
	}
	if !resp {
		t.Error("did not get both responses")
	}
}
