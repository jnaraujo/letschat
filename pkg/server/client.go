package server

import (
	"time"

	"github.com/jnaraujo/letschat/pkg/account"
)

type Client struct {
	Conn     Connection
	Account  *account.Account
	JoinedAt time.Time
}

func NewClient(account *account.Account, conn Connection) *Client {
	return &Client{
		Account:  account,
		JoinedAt: time.Now(),
		Conn:     conn,
	}
}
