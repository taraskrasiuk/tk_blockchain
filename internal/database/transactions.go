package database

import (
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Account
type Account common.Address

func NewAccount(name string) Account {
	return Account(common.BytesToAddress([]byte(name)))
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

func (t *Tx) Hash() (Hash, error) {
	txJson, err := t.Encode()
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(txJson), nil
}

func (t *Tx) Encode() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Tx) IsReward() bool {
	return t.Data == "reward"
}
