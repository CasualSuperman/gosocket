package gosocket_test

import (
	"fmt"
	"net/http"

	gs "github.com/CasualSuperman/gosocket"
)

func ExampleServer_Handle() {
	s := gs.NewServer()
	s.Handle("/hello", printHandler)
	s.Handle("/hello", responseHandler)

	http.Handle("/gs/", s)
	http.HandleFunc("/", index)
	http.ListenAndServe(":6060", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(page))
}

func printHandler(msg gs.Msg) {
	var str string

	err := msg.Receive(&str)

	if err != nil {
		fmt.Println(err.Error())
		return
	}


	fmt.Println("Hello, " + str + "!")
}

func responseHandler(msg gs.Msg) {
	var str string

	err := msg.Receive(&str)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	msg.Respond("Hello, " + str + "!")
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

	gs.Send("/hello", "world").response(function(msg) {
		document.getElementById("log").appendChild(document.createTextNode("Server says: " + msg + "\n"));
	});

	setTimeout(gs.Close, 1000);
	</script>
</body>
</html>`
