package account

import (
	"github.com/jnaraujo/letschat/pkg/id"
)

type Account struct {
	ID       id.ID  `json:"id"`
	Username string `json:"username"`
}

func NewAccount(username string) *Account {
	return &Account{
		ID:       id.NewID(22),
		Username: username,
	}
}
