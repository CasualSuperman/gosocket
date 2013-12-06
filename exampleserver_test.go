package gosocket_test

import (
	"net/http"

	gs "github.com/CasualSuperman/gosocket"
)

func ExampleServer() {
	s := gs.NewServer()

	s.On(gs.Connect, func(c *gs.Conn) {
		c.Send("hello", "world")
	})

	http.ListenAndServe(":6060", s)
}

/*
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

func index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(page))
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

	setTimeout(gs.close, 2000);
	</script>
</body>
</html>`
*/
