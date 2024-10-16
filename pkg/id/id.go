package id

import (
	"github.com/jnaraujo/letschat/pkg/secure"
)

type ID string

func NewID(n int) ID {
	return ID(secure.GenerateRandomString(n))
}
