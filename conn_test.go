package gosocket

import (
	"net/http"
	"testing"
	"time"
	"sync"
)

func TestOpen(t *testing.T) {
	var wg sync.WaitGroup
	connected := false
	s := NewServer()

	wg.Add(1)
	s.On(Connect, func(c *Conn) {
		connected = true
		wg.Done()
	})

	go http.ListenAndServe(":6669", s)

	// Give the server time to bind.
	time.Sleep(100 * time.Millisecond)

	_, err :=Open("localhost:6669")
	if err != nil {
		t.Error(err)
	}

	wg.Wait()
	if !connected {
		t.Error("failed to detect connection")
	}
}

func TestClose(t *testing.T) {

}
