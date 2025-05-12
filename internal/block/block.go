package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"taraskrasiuk/blockchain_l/internal/transactions"
	"time"
)

type Hash [32]byte

func (h *Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

type Block struct {
	Header  BlockHeader       `json:"header"`
	Payload []transactions.Tx `json:"payload"`
}

type BlockFS struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

func NewBlock(parentHash Hash, num uint64, payload []transactions.Tx) Block {
	h := BlockHeader{
		ParentHash: parentHash,
		Number:     num,
		Time:       uint64(time.Now().Unix()),
	}
	return Block{
		Header:  h,
		Payload: payload,
	}
}

type BlockHeader struct {
	ParentHash Hash   `json:"parentHash"`
	Number     uint64 `json:"number"`
	Time       uint64 `json:"time"`
}

func (b Block) Hash() (Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(blockJson), nil
}
