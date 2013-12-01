package main

import (
	"fmt"
	"net/http"

	gs "github.com/CasualSuperman/gosocket"
)

func main() {
	s := gs.NewServer()
	s.Handle("/hello", helloHandler)
	s.Handle("/goodbye", goodbyeHandler)
	s.Closed(func(c gs.Conn) {
		fmt.Println("Connection closed.")
	})
	s.Errored(func(err error) {
		fmt.Println("Error encountered:", err.Error())
	})

	http.Handle("/gs/", s)
	http.HandleFunc("/", index)
	http.ListenAndServe(":6060", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(page))
}

func helloHandler(c gs.Conn, d gs.Data) {
	var msg string

	err := d.Receive(&msg)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	str := "Hello, " + msg + "!"

	fmt.Println(str)
	c.Send("/say", str)
}

func goodbyeHandler(c gs.Conn, d gs.Data) {
	var msg string

	err := d.Receive(&msg)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	str := "Goodbye, " + msg + "."

	fmt.Println(str)
	c.Send("/say", str)
}

const page = `<!DOCTYPE html>
<html>
<head>
	<title>WebChan</title>
	<script src="/gs/gs.js"></script>
</head>
<body>
<pre id="log"></pre>
	<script>
	var gs = new GoSocket("localhost:6060/gs/");
	gs.On("/say", function(msg) {
		document.getElementById("log").appendChild(document.createTextNode("Server says: " + msg + "\n"));
	});
	gs.Ready(function() {
		gs.Send("/hello", "world");
		gs.Send("/goodbye", "cruel world");
		gs._conn.send("error");
	})
	setTimeout(gs.Close, 1000);
	</script>
</body>
</html>`
