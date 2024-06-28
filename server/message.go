package server

import (
	"encoding/json"
	"log"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const UserJoinedAction = "user-joined"
const UserLeftAction = "user-left"
const RoomJoinedAction = "room-joined"

type Message struct {
	Sender  *Client `json:"sender"`
	Target  *Room   `json:"target"`
	Action  string  `json:"action"`
	Message string  `json:"message"`
}

func (message *Message) Encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
