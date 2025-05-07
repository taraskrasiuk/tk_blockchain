package transactions

import (
	"time"
)

// Account
type Account string

func NewAccount(name string) Account {
	return Account(name)
}

// Transaction
type Tx struct {
	From      Account `json:"from"`
	To        Account `json:"to"`
	Value     uint    `json:"value"`
	Data      string  `json:"data"`
	CreatedAt string  `json:"createdAt"`
}

func NewTx(from, to Account, data string, value uint) *Tx {
	createdAt := time.Now().Format(time.RFC3339)

	return &Tx{from, to, value, data, createdAt}
}

func (t *Tx) IsReward() bool {
	return t.Data == "reward"
}
