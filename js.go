package gosocket

const js = `(function(window) {
"use strict";

/** @constructor */
function Msg(socketMsg) {
	var data = JSON.parse(socketMsg.data);

	this.data = JSON.parse(data["Msg"]);
	this.isResponse = data["IsResp"];
	this.id = data["ID"];
	this.path = data["Path"];

	return this;
}

/** @constructor */
function GoSocket(url) {
	var conn = new WebSocket("ws://" + url),
	    paths = {},
	    conversations = {},
	    id = Math.floor(Math.random() * 100000000),
	    sendqueue = [];

	function addResponse (id, ret) {
		return function(cb) {
			var cbs = conversations[id];
			if (!cbs) {
				cbs = [];
			}
			cbs.push(cb);
			conversations[id] = cbs;

			return ret;
		}
	}

	conn.onopen = function() {
		while(sendqueue.length > 0) {
			conn.send(sendqueue.shift());
		}
	};

	conn.onmessage = function(resp) {
		var msg = new Msg(resp),
		    pass = {data: msg.data};

		pass.response = addResponse(msg.id, pass);

		pass.respond = function(data) {
			data = JSON.stringify(data);

			conn.send(JSON.stringify({
				"Msg": data,
				"ID": msg.id,
				"IsResp": true
			}));

			return pass;
		};

		if (msg.isResponse) {
			var cbs = conversations[msg.id];
			if (cbs) {
				var cb = cbs.shift();
				cb(pass);
				if (cbs.length === 0) {
					cbs = undefined;
				}
				conversations[msg.id] = cbs;
			}
		} else {
			var handlers = paths[msg.path];

			if (handlers) {
				for (var i = handlers.length-1; i >= 0; i--) {
					handlers[i](pass);
				}
			}
		}
	};

	this.send = function(path, msg) {
		var msg = JSON.stringify(msg),
		    data = JSON.stringify({
		    	"Path": path,
		    	"Msg": msg,
		    	"ID": id}),
		    ret = {};

		if (conn.readyState >= conn.OPEN) {
			conn.send(data);
		} else {
			sendqueue.push(data);
		}

		ret.response = addResponse(id, ret);

		id++;

		return ret;
	};

	this.on = function(path, func) {
		var handlers = paths[path];
		if (!handlers) {
			handlers = [];
		}
		handlers.push(func);
		paths[path] = handlers;
	};

	this.close = function() {
		conn.close()
	};

	this._conn = conn;

	return this;
};

window["GoSocket"] = GoSocket;
}(window))`
