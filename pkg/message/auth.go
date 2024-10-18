package message

import "github.com/jnaraujo/letschat/pkg/id"

type AuthMessageClient struct {
	Username string `json:"username"`
	RoomID   id.ID  `json:"room_id"`
	// Add stuff here like public key, ID, etc.
}

type AuthMessageServer struct {
	Status  string `json:"status"`
	Content string `json:"content"`
	RoomID  id.ID  `json:"room_id"`
}
