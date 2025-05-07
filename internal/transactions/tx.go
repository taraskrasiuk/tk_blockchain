package transactions

// Account
type Account string

// Transaction
type Tx struct {
	From      Account `json:"from"`
	To        Account `json:"to"`
	Value     uint    `json:"value"`
	Data      string  `json:"data"`
	CreatedAt string  `json:"createdAt"`
}

func (t *Tx) IsReward() bool {
	return t.Data == "reward"
}
