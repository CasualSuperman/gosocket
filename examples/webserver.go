package main

import (
	"fmt"
	"net/http"

	wc "github.com/CasualSuperman/webchan"
)

func main() {
	s := wc.NewServer()
	s.Handle("/hello", helloHandler)
	s.Handle("/goodbye", goodbyeHandler)
	s.Closed(func(c wc.Conn) {
		fmt.Println("Connection closed.")
	})
	s.Errored(func(err error) {
		fmt.Println("Error encountered:", err.Error())
	})

	http.Handle("/wc", s)
	http.HandleFunc("/wc.js", wc.JavaScript)
	http.HandleFunc("/", index)
	http.ListenAndServe(":6060", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(page))
}

func helloHandler(c wc.Conn, d wc.Data) {
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

func goodbyeHandler(c wc.Conn, d wc.Data) {
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
	<script src="http://localhost:6060/wc.js"></script>
</head>
<body>
<pre id="log"></pre>
	<script>
	var wc = new WebChan("localhost:6060/wc");
	wc.On("/say", function(msg) {
		document.getElementById("log").appendChild(document.createTextNode("Server says: " + msg + "\n"));
	});
	wc.Ready(function() {
		wc.Send("/hello", "world");
		wc.Send("/goodbye", "cruel world");
		wc._conn.send("error");
	})
	setTimeout(wc.Close, 1000);
	</script>
</body>
</html>`
