package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/message"
)

type Room struct {
	ID      id.ID
	Name    string
	Owner   *account.Account
	Clients *ClientList
}

func NewRoom(name string, owner *account.Account) *Room {
	return &Room{
		ID:      id.NewID(22),
		Name:    name,
		Owner:   owner,
		Clients: NewClientList(),
	}
}

func (r *Room) AddClient(client *Client) {
	client.RoomID = r.ID
	r.Clients.Add(client)

	r.Broadcast(
		message.NewServerChatMessage(
			fmt.Sprintf(
				"%s (%s) joined the chat", client.Account.Username, client.Account.ID,
			),
			message.CharRoom{
				ID:   r.ID,
				Name: r.Name,
			},
			time.Now(),
		),
	)
}

func (r *Room) RemoveClient(id id.ID) {
	r.Clients.Remove(id)
}

func (r *Room) HasClient(id id.ID) bool {
	return r.Clients.Has(id)
}

func (r *Room) Broadcast(msg any) {
	for _, client := range r.Clients.List() {
		client.Conn.WriteMessage(msg)
	}
}

type RoomList struct {
	rooms map[id.ID]*Room
	mutex sync.RWMutex
}

func NewRoomList() *RoomList {
	return &RoomList{
		rooms: make(map[id.ID]*Room),
	}
}

func (rl *RoomList) Add(room *Room) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.rooms[room.ID] = room
}

func (rl *RoomList) Find(id id.ID) *Room {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	return rl.rooms[id]
}

func (rl *RoomList) Remove(id id.ID) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	delete(rl.rooms, id)
}

func (rl *RoomList) Has(id id.ID) bool {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	_, exists := rl.rooms[id]
	return exists
}

func (rl *RoomList) List() []*Room {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	rooms := make([]*Room, 0, len(rl.rooms))
	for _, room := range rl.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
