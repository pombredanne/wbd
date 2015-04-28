package web

// TODO: Clean up the Javascript (wrap bits in separate template consts)

const (
	indexTemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title> Wallboard Control </title>

	<style type='text/css'>
	/* Remove padding around iframe */
	html, body {
		margin: 0;
		height: 100%;
		overflow: hidden;
	}

	iframe {
		border: 0;
		width: 100%;
		height: 100%;
	}
	</style>

	<script type='text/javascript'>
	var attempts = 1;

	Array.prototype.equals = function (array) {
		// if the other array is a falsy value, return
		if (!array)
			return false;

		// compare lengths - can save a lot of time
		if (this.length != array.length)
			return false;

		for (var i = 0, l=this.length; i < l; i++) {
			// Check if we have nested arrays
			if (this[i] instanceof Array && array[i] instanceof Array) {
				// recurse into the nested arrays
				if (!this[i].equals(array[i]))
					return false;
			}
			else if (this[i] != array[i]) {
				// Warning - two different object instances will never be equal: {x:20} != {x:20}
				return false;
			}
		}
		return true;
	}

	function generateInterval(k) {
		var maxInterval = (Math.pow(2, k) - 1) * 1000;

		if (maxInterval > 30*1000) {
			maxInterval = 30*1000; // If the generated interval is more than 30 seconds, truncate it down to 30 seconds.
		}

		// generate the interval to a random number between 0 and the maxInterval determined from above
		return Math.random() * maxInterval;
	}

	function SiteRotator (elementId, defaultUrls, duration) {
		if (typeof elementId === 'undefined') return;
		if (typeof defaultUrls === 'undefined' || defaultUrls.length < 1) return;

		this.elementId = elementId;
		this.defaultUrls = defaultUrls;
		this.duration = duration;

		this.urls = defaultUrls;
		this.currentIndex = 0;

		this.init = function() {
			// Load first URL when initialized
			console.log("Initializing rotator");

			// Try to use the default URLs if we don't have any
			if (this.urls.length < 1) {
				if (this.defaultUrls.length < 1) {
					if (typeof this.interval !== 'undefined') {
						clearInterval(this.interval);
					}

					console.error("Can't run rotator -- no URLs to rotate");
					return;
				}

				this.urls = this.defaultUrls;
			}

			this.currentIndex = 0;
			this.load(this.urls[this.currentIndex]);

			// If a duration is passed in, setup rotation
			if (typeof this.duration !== 'undefined')
			{
				this.rotateEvery(this.duration);
			}
		};

		this.setUrls = function(urls) {
			if (urls === null || urls === undefined) {
				urls = [];
			}

			// Only update if URLs changed -- this reinits the rotator
			if (this.urls.equals(urls) === false) {
				console.log("Current URLs:", this.urls);
				console.log("Updated URLs:", urls);

				this.urls = urls;
				this.init();
			}
		};

		this.load = function(url) {
			console.log("Loading URL:", url)

			document.getElementById(this.elementId).src = url;
		};

		this.next = function() {
			console.log("Moving to next URL");

			if (this.urls.length < 1) {
				return;
			}

			this.currentIndex++;
			if (this.currentIndex >= this.urls.length) {
				this.currentIndex = 0;
			}

			this.load(this.urls[this.currentIndex]);
		};

		this.previous = function() {
			if (this.urls.length < 1) {
				return;
			}

			this.currentIndex--;
			if (this.currentIndex < 0) {
				this.currentIndex = this.urls.length - 1;
			}

			this.load(this.urls[this.currentIndex]);
		};

		this.pause = function(duration) {
			// If we're already paused, bail
			if (typeof this._load !== 'undefined') { return; }

			// Save this.load and noop it
			this._load = this.load;
			this.load = function(url) { }

			setInterval(function() {
				this.resume();
			}, duration * 1000);
		}

		this.resume = function() {
			if (typeof this._load === 'undefined') { return; }

			// Return the function
			this.load = this._load;

			// Load the page we should be on
			this.load(this.urls[this.currentIndex]);
		}

		this.rotateEvery = function(duration) {
			if (typeof this.interval !== 'undefined') {
				clearInterval(this.interval);
			}

			var t = this;
			this.interval = setInterval(function() {
				t.next();
			}, duration * 1000);

			console.log("Rotate scheduled for every ", duration, "seconds");
		};

		this.init();
	}

	function wbcConnect(endpoint, rotator) {
		if (typeof endpoint === 'undefined') return false;
		if (!window["WebSocket"]) return false;

		conn = new WebSocket(endpoint);
		conn.onopen = function(evt) {
			console.log("Connected to websocket server");
			conn.send(JSON.stringify({
				"action": "sendUrls"
			}));

			// reset reconnection counter
			attempts = 1;
		}
		conn.onclose = function(evt) {
			console.log("Disconnected from websocket server");

			// attempt reconnection
			var time = generateInterval(attempts);
			console.log("Attempting reconnection in " + time + " milliseconds")

			setTimeout(function() {
				console.log("Attempting reconnection");

				attempts++;

				wbcConnect(endpoint, rotator);
			}, time);
		}
		conn.onmessage = function(evt) {
			message = JSON.parse(evt.data);

			if (typeof message.action === 'undefined')
			{
				console.error("No action in message from server:", message)
				return;
			}

			switch (message.action) {
			case 'updateUrls':
				rotator.setUrls(message.data.urls);

				break;
			case 'flashUrl':
				rotator.pause(message.data.url);
				rotator.load(message.data.duration);

				break;
			default:
				console.error("Unknown action in message from server:", message)
				break;
			}
		}
	}

	document.addEventListener("DOMContentLoaded", function(event) {
		var defaultUrls = [
			'{{ .DefaultUrl }}'
		];

		var rotator = new SiteRotator('frame', defaultUrls, 60);

		{{ if ne .Client "" }}
		// Connect to WebSocket server (provides control)
		wbcConnect("ws://{{ .Address }}/ws?client={{ .Client }}", rotator);
		{{ else }}
		wbcConnect("ws://{{ .Address }}/ws", rotator);
		{{ end }}
	});
	</script>
</head>
<body>
	<iframe id='frame'>Oops, something went wrong with the Wallboard page!</iframe>
</body>
</html>
`

	welcomeTemplate string = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title> Welcome </title>
	<style type='text/css'>
	html, body {
		height: 100%;
		width: 100%;
		background-color: #000;
		color: #fff;
		margin: 0;
		padding: 0;
	}
	div.wrapper {
		position: absolute;
		left: 50%;
		top: 50%;
		transform: translate(-50%, -50%);
		-webkit-transform: translate(-50%, -50%);
		-moz-transform: translate(-50%, -50%);
		-ms-transform: translate(-50%, -50%);
	}
	h1 {
		font-size: 6em;
	}
	</style>
</head>
<body>
	<div class='wrapper'>
		<h1>wbc</h1>
		{{ if ne .Client "" }}<h2>Client: {{ .Client }}</h2>{{end}}
		<h2>IP Addr: {{ .RemoteAddr }}</h2>
		<p>Add a URL or two and this page will disappear. :)</p>
	</div>
</body>
</html>
`

	adminTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title> Wallboard Control </title>

	<style type='text/css'>
	/* Remove padding around iframe */
	html, body {
		margin: 0;
		height: 100%;
		overflow: hidden;
	}

	div#wrapper {
		height: 100%;
		width: 100%;
	}

	div#output-box {
		overflow-y: scroll;
		overflow-x: hidden;
		height: 94%;
		width: 100%;
	}

	div#input-box {
		width: 100%;
	}
	input#input {
		width: 100%;
	}

	div.message {
		padding: 3px;
		clear: both;
		border-bottom: 1px solid #fff;
	}

	div.output-message {
		background-color: #375EAB;
		color: #fff;
		width: 100%;
	}

	div.input-message {
		background-color: #E0EBF5;
		color: #000;
		width: 100%;
	}
	</style>

	<script type='text/javascript' src='https://code.jquery.com/jquery-2.1.3.min.js'></script>
	<script type='text/javascript'>
	String.prototype.lpad = function(padString, length) {
		var str = this;
		while (str.length < length)
			str = padString + str;
		return str;
	}

	$(function() {
		var attempts = 1;

		function generateInterval(k) {
			var maxInterval = (Math.pow(2, k) - 1) * 1000;

			if (maxInterval > 30*1000) {
				maxInterval = 30*1000; // If the generated interval is more than 30 seconds, truncate it down to 30 seconds.
			}

			// generate the interval to a random number between 0 and the maxInterval determined from above
			return Math.random() * maxInterval;
		}

		function wbcConnect(endpoint, inputElement, outputElement) {
			if (typeof endpoint === 'undefined') return false;
			if (typeof inputElement === 'undefined') return false;
			if (typeof outputElement === 'undefined') return false;
			if (!window["WebSocket"]) return false;

			this.endpoint = endpoint;
			this.inputElement = inputElement;
			this.outputElement = outputElement;
			this.lastUrls = null;

			var getTime = function() {
				date = new Date();
				strDate = date.getHours().toString().lpad("0", 2) + ":" +
				          date.getMinutes().toString().lpad("0", 2) + ":" +
				          date.getSeconds().toString().lpad("0", 2);

				return strDate;
			}

			var print = function(msg, type) {
				timestamp = getTime()
				$("<div class='message " + type + "-message'>[" + timestamp + "] " + msg + "</div>").appendTo(this.outputElement);
				this.outputElement.scrollTop(this.outputElement.prop("scrollHeight"));
			};

			// Create the WebSocket
			this.conn = new WebSocket(endpoint);

			var self = this;

			this.conn.onopen = function(evt) {
				console.log("Connected to websocket server");
				print("Connected to server!", "generic")

				// reset connection attempt counter
				attempts = 1;

				conn.send(JSON.stringify({
					"action": "flagController"
				}));
			}

			this.conn.onclose = function(evt) {
				console.log("Disconnected from websocket server");

				// Don't print disconnection message on every reconnection
				// attempt
				if (attempts == 1)
				{
					print("Disconnected from server!", "generic")
				}

				// attempt reconnection
				var time = generateInterval(attempts);
				console.log("Attempting reconnection in " + time + " milliseconds")

				setTimeout(function() {
					console.log("Attempting reconnection");

					attempts++;

					wbcConnect(endpoint, inputElement, outputElement);
				}, time);
			}

			this.conn.onmessage = function(evt) {
				message = JSON.parse(evt.data);

				if (typeof message.action === 'undefined')
				{
					console.error("No action in message from server:", message)
					return;
				}

				switch (message.action) {
				case 'updateUrls':
					if (message.data.urls != self.lastUrls)
					{
						if (message.data.urls == null) {
							break;
						}

						strData = JSON.stringify(message.data);

						if (self.lastUrls == null) {
							print("URL list received: " + strData, 'output');
						} else {
							print("Updated URL list received: " + strData, 'output');
						}

						self.lastUrls = message.data.urls;
					}

					break;
				case 'updateClients':
					if (message.data.clients != self.lastClients)
					{
						if (message.data.clients == null)
						{
							break;
						}

						strData = JSON.stringify(message.data);

						if (self.lastClients == null) {
							print("List of clients received: " + strData, 'output');
						} else {
							print("Updated list of clients received: " + strData, 'output');
						}

						self.lastClients = message.data.clients;
					}

					break;
				default:
					console.error("Unknown action in message from server:", message)
					break;
				}
			}

			// Handler for input box
			this.inputElement.keyup(function(evt){
				if(evt.keyCode == 13) {
					switch ($(this).val()) {
						default:
							message = JSON.stringify({
								action: $(this).val()
							});

							print($(this).val() + ' -> ' + message, "input")

							try {
								self.conn.send(message);
							} catch(e) {
								print("Unable to send message to server", "generic");
							}
						break;
					}
					$(this).val("");
				}
			});
		}

		wbcConnect(
			{{ if ne .Client "" }}
			// Connect to WebSocket server (provides control)
			"ws://{{ .Address }}/ws?client={{ .Client }}",
			{{ else }}
			"ws://{{ .Address }}/ws",
			{{ end }}
			$('#input'),
			$('#output-box')
		);

		$('#input').focus();
	});
	</script>
</head>
<body>
	<div id='wrapper'>
		<div id='output-box'>
		</div>
		<div id='input-box'>
			<input type='text' id='input'>
		</div>
	</div>
</body>
</html>
`
)
