package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Hash [32]byte

func (h *Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

type Block struct {
	Header  BlockHeader `json:"header"`
	Payload []SignedTx  `json:"payload"`
}

func NewBlock(parentHash Hash, num uint64, nonce uint32, payload []SignedTx, miner common.Address) Block {
	h := BlockHeader{
		ParentHash: parentHash,
		Number:     num,
		Time:       uint64(time.Now().Unix()),
		Nonce:      nonce,
		Miner:      miner,
	}
	return Block{
		Header:  h,
		Payload: payload,
	}
}

type BlockHeader struct {
	ParentHash Hash           `json:"parentHash"`
	Number     uint64         `json:"number"`
	Nonce      uint32         `json:"nonce"`
	Time       uint64         `json:"time"`
	Miner      common.Address `json:"miner"`
}

func (b Block) Hash() (Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return Hash{}, err
	}
	return sha256.Sum256(blockJson), nil
}

type BlockFS struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

// Block validation
func IsValidBlock(h Hash) bool {
	return fmt.Sprintf("%x", h[0]) == "0" &&
		fmt.Sprintf("%x", h[1]) == "0" &&
		// fmt.Sprintf("%x", h[2]) == "0" &&
		// fmt.Sprintf("%x", h[3]) == "0" &&
		// not equal to zero
		fmt.Sprintf("%x", h[2]) != "0"
}
