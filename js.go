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
		};
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

	this.send = function(path, m) {
		var msg = JSON.stringify(m),
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
		conn.close();
	};

	this._conn = conn;

	return this;
}

window["GoSocket"] = GoSocket;
}(window));`

const jsMin = `(function(h){function n(e){e=JSON.parse(e.data);this.data=JSON.parse(e.Msg);this.a=e.IsResp;this.id=e.ID;this.path=e.Path;return this}h.GoSocket=function(e){function m(a,b){return function(d){var c=g[a];c||(c=[]);c.push(d);g[a]=c;return b}}var f=new WebSocket("ws://"+e),h={},g={},k=Math.floor(1E8*Math.random()),l=[];f.onopen=function(){for(;0<l.length;)f.send(l.shift())};f.onmessage=function(a){var b=new n(a),d={data:b.data};d.response=m(b.id,d);d.b=function(a){a=JSON.stringify(a);f.send(JSON.stringify({Msg:a,
ID:b.id,IsResp:!0}));return d};if(b.a){if(a=g[b.id])a.shift()(d),0===a.length&&(a=void 0),g[b.id]=a}else if(a=h[b.path])for(var c=a.length-1;0<=c;c--)a[c](d)};this.send=function(a,b){b=JSON.stringify(b);var d=JSON.stringify({Path:a,Msg:b,ID:k}),c={};f.readyState>=f.OPEN?f.send(d):l.push(d);c.response=m(k,c);k++;return c};this.close=function(){f.close()};return this}})(window);`
