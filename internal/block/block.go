package block

import (
	"crypto/sha256"
	"encoding/json"
	"taraskrasiuk/blockchain_l/internal/transactions"
)

type Hash [32]byte

type Block struct {
	Header  BlockHeader
	Payload []transactions.Tx
}

type BlockHeader struct {
	ParentHash Hash
	Time       uint64
}

func (b *Block) Hash() (Hash, error) {
	d, err := json.Marshal(b)
	if err != nil {
		return *new(Hash), err
	}
	return sha256.Sum256(d), nil
}
