package database

import (
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Account
func NewAccount(name string) common.Address {
	return common.HexToAddress(name)
}

// Transaction
type Tx struct {
	From      common.Address `json:"from"`
	To        common.Address `json:"to"`
	Value     uint           `json:"value"`
	Data      string         `json:"data"`
	CreatedAt string         `json:"createdAt"`
}

func NewTx(from, to common.Address, data string, value uint) *Tx {
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

// Signed Transaction
type SignedTx struct {
	Tx
	Sig []byte `json:"signature"`
}

func NewSignedTx(tx Tx, sig []byte) *SignedTx {
	return &SignedTx{tx, sig}
}

func (t *SignedTx) IsAuthentic() (bool, error) {
	txHash, err := t.Tx.Hash()
	if err != nil {
		return false, err
	}
	recoveredPubKey, err := crypto.SigToPub(txHash[:], t.Sig)
	if err != nil {
		return false, err
	}
	recoveredPubKeyBytes := elliptic.Marshal(
		crypto.S256(),
		recoveredPubKey.X,
		recoveredPubKey.Y,
	)
	recoveredPubKeyBytesHash := crypto.Keccak256(
		recoveredPubKeyBytes[1:],
	)
	recoveredAccount := common.BytesToAddress(
		recoveredPubKeyBytesHash[12:],
	)
	return recoveredAccount.Hex() == common.Address(t.Tx.From).Hex(), nil
}
