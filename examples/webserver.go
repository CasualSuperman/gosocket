package main

import (
	"fmt"
	"net/http"

	gs "github.com/CasualSuperman/GoSocket"
)

func main() {
	s := gs.NewServer()
	s.Handle("/hello", helloHandler)
	s.Handle("/goodbye", goodbyeHandler)
	s.Handle("/conversation", convoHandler)
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

func helloHandler(msg gs.Msg) {
	var target string

	err := msg.Receive(&target)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	str := "Hello, " + target + "!"

	fmt.Println(str)
	msg.Respond(str)
}

func goodbyeHandler(msg gs.Msg) {
	var target string

	err := msg.Receive(&target)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	str := "Goodbye, " + target + "."

	fmt.Println(str)
	msg.Respond(str)
}

func convoHandler(msg gs.Msg) {
	var target string

	err := msg.Receive(&target)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Convo: got", target)

	err = msg.Respond(target + ", ducks")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Convo: sent ducks")

	msg, err = msg.Response()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = msg.Receive(&target)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Convo: got", target)

	msg.Respond(target + ", turtles")

	fmt.Println("Convo: sent turtles")
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

	function log(msg) {
		document.getElementById("log").appendChild(document.createTextNode("Server says: " + msg + "\n"));
	}

	gs.send("/hello", "world").response(function(msg) {
		log(msg.data);
	});

	gs.send("/conversation", "world").response(function(msg) {
		console.log("Got " + msg.data + ".");
		console.log("Sent geese");
		msg.respond(msg.data + ", geese");
	}).response(function(msg) {
		console.log("Got " + msg.data + ".");
		log(msg.data);
	});

	gs.send("/goodbye", "cruel world").response(function(msg) {
		log(msg.data);
	});

	setTimeout(gs.close, 2000);
	</script>
</body>
</html>`
