package database

import (
	"encoding/hex"
	"reflect"
	"testing"
	"time"
)

func createTx(from, to string, value uint) SignedTx {
	return SignedTx{
		Tx: Tx{
			From:      NewAccount(from),
			To:        NewAccount(to),
			Value:     value,
			CreatedAt: time.Now().Format(time.RFC3339),
		},
	}
}

func TestNewBlock(t *testing.T) {
	parentHash := Hash{}
	txs := []SignedTx{
		createTx("from", "to", 100),
	}
	block0 := NewBlock(parentHash, 1, 0x0123, txs, NewAccount("miner"))
	bhash0, err := block0.Hash()
	if err != nil {
		t.Fatalf("got an error %v", err)
	}
	// if !reflect.DeepEqual(block0.Header.ParentHash, Hash{}) {
	// 	t.Fatalf("the parent hash is not correct, expected %x but got %x", Hash{}, block0.Header.ParentHash)
	// }
	block1 := NewBlock(bhash0, 2, 0x0000123, append(txs, []SignedTx{createTx("from-2", "to2", 200)}...), NewAccount("miner"))
	_, err = block1.Hash()
	if err != nil {
		t.Fatalf("got an error %v", err)
	}
	if !reflect.DeepEqual(block1.Header.ParentHash, bhash0) {
		t.Fatal("expected a parent hash to be equal to block 0 hash")
	}
}

func TestValidBlock(t *testing.T) {
	nonce := "00000000fa04f816039...a4db586086168edfa"
	var hash = Hash{}
	hex.Decode(hash[:], []byte(nonce))

	if !IsValidBlock(hash) {
		t.Fatal("the hash is not correct")
	}
}
