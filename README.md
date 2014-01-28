GoSocket [![SemVer](http://img.shields.io/semver/v0.6.3.png?color=yellow)](http://semver.org/ "SemVer") [![Build Status](https://travis-ci.org/CasualSuperman/gosocket.png)](https://travis-ci.org/CasualSuperman/gosocket) [![GoDoc](http://godoc.org/github.com/CasualSuperman/gosocket?status.png)](http://godoc.org/github.com/CasualSuperman/gosocket) [![CodeBot](http://img.shields.io/codebot/A+.png)](http://codebot.io/doc/pkg/github.com/CasualSuperman/gosocket "CodeBot") [![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/CasualSuperman/gosocket/trend.png)](https://bitdeli.com/free "Bitdeli Badge")
========

GoSocket is a library that runs on top of websockets and provides a
[Socket.IO](http://socket.io/)-like interface for communicating with the
browser or other Go programs.

GoSocket was initially developed for use within
[Diorite](https://github.com/CasualSuperman/Diorite).  I needed a way to
quickly transfer data between the main application and the web UI, and I
initially turned to websockets.  Unfortunately, websockets required an extra
layer of abstraction to differentiate between API locations.  Websocket servers
can run on only a single path each within an http server, so my idea of setting
up multiple paths on the server wasn't feasible.  I designed GoSockets to
internalize these paths, allowing for a single websocket connection to simulate
multiple endpoints.
