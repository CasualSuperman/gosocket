package gosocket

import (
	"net/http"
	"testing"
	"time"
)

var msgServer *Server

func init() {
	msgServer = startMsgTestServer()
}

func startMsgTestServer() *Server {
	s := NewServer()
	go func() {
		err := http.ListenAndServe(":6667", s)
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	return s
}

func connectMsgTestServer() (*Conn, error) {
	return Open("localhost:6667")
}

func clearMsgServer() {
	clearServer(msgServer)
}

func TestMsgReceive(t *testing.T) {
	type testStruct struct {
		Field1, Field2 int
		Field3 string
	}
	defer clearMsgServer()
	done := make(chan bool)
	c, _ := connectMsgTestServer()

	msgServer.Handle("string", func(m Msg) {
		var s string
		m.Receive(&s)
		if s != "world" {
			t.Error("receiving string failed")
		}
		done <- true
	})
	c.Send("string", "world")
	<-done

	msgServer.Handle("struct", func(m Msg) {
		var ts testStruct
		err := m.Receive(&ts)
		if err != nil {
			t.Error("receiving struct failed")
		}
		if ts.Field1 != 1 || ts.Field2 != 2 || ts.Field3 != "test" {
			t.Error("struct values incorrect")
		}
		done <- true
	})
	c.Send("struct", testStruct{1,2,"test"})
	<-done

	msgServer.Handle("int", func(m Msg) {
		var i int
		err := m.Receive(&i)
		if err != nil {
			t.Error("receiving int failed")
		}
		if i != 42 {
			t.Error("int value incorrect")
		}
		done <- true
	})
	c.Send("int", 42)
	<-done
}

func TestMsgRespond(t *testing.T) {
	defer clearMsgServer()
	done := make(chan bool)
	c, _ := connectMsgTestServer()

	msgServer.Handle("echo", func(m Msg) {
		var s string
		m.Receive(&s)
		err := m.Respond(s)
		if err != nil {
			t.Error("respond failed: " + err.Error())
		}
		done <- true
	})

	c.Send("echo", "Alpha")
	<-done
}

func TestMsgRespondFail(t *testing.T) {
	defer clearMsgServer()
	done := make(chan bool)
	c, _ := connectMsgTestServer()

	msgServer.Handle("echo", func(m Msg) {
		err := m.Respond(func(){})
		if err == nil {
			t.Error("respond should not have succeeded")
		}
		done <- true
	})

	c.Send("echo", "Alpha")
	<-done
}

func TestMsgResponse(t *testing.T) {
	defer clearMsgServer()
	c, _ := connectMsgTestServer()

	msgServer.Handle("echo", func(m Msg) {
		var s string
		m.Receive(&s)
		m.Respond(s)
	})

	msg, _ := c.Send("echo", "Alpha")
	msg2, err := msg.Response()
	if err != nil {
		t.Error("could not get response")
		return
	}
	var s string
	msg2.Receive(&s)
	if s != "Alpha" {
		t.Error("response did not match")
	}
}

func TestMsgResponseAlreadyClosed(t *testing.T) {
	defer clearMsgServer()
	done := make(chan bool)
	c, _ := connectMsgTestServer()

	msgServer.Handle("echo", func(m Msg) {
		done <- true
	})

	msg, _ := c.Send("echo", "Alpha")
	<-done
	c.Close()
	_, err := msg.Response()
	if err == nil {
		t.Error("response should fail")
	}
}

func TestMsgResponseClosedAfter(t *testing.T) {
	defer clearMsgServer()
	done := make(chan bool)

	msgServer.On(Connect, func(c *Conn) {
		done <- true
	})

	c, _ := connectMsgTestServer()

	go func() {
		<-done
		time.Sleep(100 * time.Millisecond)
		c.Close()
	}()


	msg, _ := c.Send("test", func(){})
	_, err := msg.Response()
	if err == nil {
		t.Error("response should fail")
	}
}

func TestMsgTimedResponseTimeout(t *testing.T) {
	defer clearMsgServer()

	c, _ := connectMsgTestServer()
	msg, _ := c.Send("test", nil)
	msg, err := msg.TimedResponse(10 * time.Millisecond)

	if err == nil || err.Error() != "operation timed out" {
		t.Error("response should have timed out")
	}
}

func TestMsgTimedResponseReceived(t *testing.T) {
	defer clearMsgServer()

	msgServer.Handle("test", func(m Msg) {
		m.Respond(nil)
	})

	c, _ := connectMsgTestServer()
	msg, _ := c.Send("test", nil)
	msg, err := msg.TimedResponse(10 * time.Millisecond)

	if err != nil {
		t.Error("response should not have timed out")
	}
}
