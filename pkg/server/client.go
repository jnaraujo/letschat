package server

import (
	"sync"
	"time"

	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
)

type Client struct {
	Conn     Connection
	Account  *account.Account
	RoomID   id.ID
	JoinedAt time.Time
}

func NewClient(account *account.Account, conn Connection) *Client {
	return &Client{
		Account:  account,
		JoinedAt: time.Now(),
		Conn:     conn,
	}
}

type ClientList struct {
	clients map[id.ID]*Client
	mutex   sync.RWMutex
}

func NewClientList() *ClientList {
	return &ClientList{
		clients: make(map[id.ID]*Client),
	}
}

func (cl *ClientList) Add(client *Client) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	cl.clients[client.Account.ID] = client
}

func (cl *ClientList) Find(id id.ID) *Client {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()
	return cl.clients[id]
}

func (cl *ClientList) Remove(id id.ID) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	delete(cl.clients, id)
}

func (cl *ClientList) Has(id id.ID) bool {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	_, exists := cl.clients[id]
	return exists
}

func (cl *ClientList) List() []*Client {
	cl.mutex.RLock()
	defer cl.mutex.RUnlock()

	clients := make([]*Client, 0, len(cl.clients))
	for _, client := range cl.clients {
		clients = append(clients, client)
	}
	return clients
}
