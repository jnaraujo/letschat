package account

import (
	"github.com/jnaraujo/letschat/pkg/id"
)

type Account struct {
	ID       id.ID
	Username string
}

func NewAccount(username string) *Account {
	return &Account{
		ID:       id.NewID(8),
		Username: username,
	}
}
