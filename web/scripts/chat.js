document.addEventListener("DOMContentLoaded", () => {
	const inputField = document.getElementById("name-input");
	const submitButton = document.getElementById("submit-button");

	inputField.addEventListener("input", () => {
		if (inputField.value.trim() !== "") {
			submitButton.disabled = false;
		} else {
			submitButton.disabled = true;
		}
	});


	document.getElementById("home-form").addEventListener("submit", (event) => {
		event.preventDefault();
		dial();
	});
});

// Handle websocket events
function dial() {
	const name = document.getElementById("name-input").value;
	const conn = new WebSocket(`ws://${location.host}/ws?name=` + name);

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
}

// TODO: enable reconnect after a brief disconnection

// onsubmit publishes the message from the user when the form is submitted.
publishForm.onsubmit = async (ev) => {
	ev.preventDefault();

	const msg = messageInput.value;
	if (msg === "") {
		return;
	}
	messageInput.value = "";

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


(() => {
	const homeForm = document.getElementById("home-form");
	const nameInput = document.getElementById("name-input");
	const roomInput = document.getElementById("room-input");

	const messageLog = document.getElementById("message-log");
	const publishForm = document.getElementById("publish-form");
	const messageInput = document.getElementById("message-input");

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
	}

	function appendLog(text, error) {
		const p = document.createElement("p");
		// Adding a timestamp to each message makes the log easier to read.
		p.innerText = `${new Date().toLocaleTimeString()}: ${text}`;
		if (error) {
			p.style.color = "red";
			p.style.fontStyle = "bold";
		}
		messageLog.append(p);
		return p;
	}
	appendLog("Submit a message to get started!");
})();
