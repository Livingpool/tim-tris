document.addEventListener("DOMContentLoaded", () => {
	// Initialises a websocket request
	document.getElementById("home-form").addEventListener("submit", (event) => {
		if (document.getElementById("name-input").value.trim() == "") {
			alert("Enter a valid name!");
			return false;
		} else {
			dial();
		}
	});

});

// Handle websocket events
function dial() {
	const name = document.getElementById("name-input").value;
	const conn = new WebSocket(`ws://${location.host}/ws?name=` + name);
	console.log(location.host);

	conn.addEventListener("close", (event) => {
		appendLog(`WebSocket Disconnected code: ${event.code}, reason: ${event.reason}`, true);
		if (event.code !== 1001 || event.code != 1007) {
			appendLog("Reconnecting in 1s", true);
			setTimeout(dial, 1000);
		}
	});

	conn.addEventListener("open", (event) => {
		console.info("websocket connected");
	});

	// onsubmit publishes the message from the user when the form is submitted.
	document.getElementById("publish-form").onsubmit = async (event) => {
		event.preventDefault();

		const msg = messageInput.value;
		if (msg === "") {
			return;
		}
		document.getElementById("message-input").value = "";

		// TODO: handle different message actions
		expectingMessage = true;
		try {
			const jsonMessage = {
				target: "new room",
				action: "send-message",
				message: "hi cat girl",
			};
			conn.send(JSON.stringify(jsonMessage));
		} catch (err) {
			appendLog(`Publish failed: ${err.message}`, true);
		}
	}

	// appendLog appends the passed text to messageLog.
	// expectingMessage is set to true
	// if the user has just submitted a message
	// and so we should scroll the next message into view when received.
	let expectingMessage = false;
	// TODO: this is where we handle messages received.
	conn.onmessage = (event) => {
		if (typeof event.data !== "string") {
			console.error("unexpected message type", typeof ev.data);
			return;
		}
		const p = appendLog(event.data);
		if (expectingMessage) {
			p.scrollIntoView();
			expectingMessage = false;
		}
	};

	appendLog("Submit a message to get started!");
}

// TODO: enable reconnect after a brief disconnection

function appendLog(text, error) {
	const p = document.createElement("p");
	// Adding a timestamp to each message makes the log easier to read.
	p.innerText = `${new Date().toLocaleTimeString()}: ${text}`;
	if (error) {
		p.style.color = "red";
		p.style.fontStyle = "bold";
	}
	document.getElementById("message-input").append(p);
	return p;
}
