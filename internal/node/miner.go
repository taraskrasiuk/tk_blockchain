package node

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"taraskrasiuk/blockchain_l/internal/database"
	"time"
)

type PendingBlock struct {
	parent database.Hash
	number uint64
	time   uint64
	txs    []database.Tx
	miner  database.Account
}

func NewPendingBlock(h database.Hash, n uint64, txs []database.Tx, miner database.Account) *PendingBlock {
	fmt.Println("MINER: ", miner)
	if miner == "" {
		panic("The miner is an empty string")
	}
	return &PendingBlock{h, n, uint64(time.Now().UnixMilli()), txs, miner}
}

// Main Mine function
func Mine(ctx context.Context, p *PendingBlock) (database.Block, error) {
	if len(p.txs) == 0 {
		return database.Block{}, errors.New("the transactions are missed")
	}
	var (
		startTime = time.Now()
		attempt   = 0
		block     = database.Block{}
		hash      = database.Hash{}
		nonce     = *new(uint32)
	)

	// run the loop
	for !database.IsValidBlock(hash) {
		select {
		case <-ctx.Done():
			fmt.Println("Mining canceled")
			return database.Block{}, fmt.Errorf("mining canceled, %s", ctx.Err())
		default:
		}

		attempt++
		nonce = generateNonce()

		if attempt%1_000_000 == 0 || attempt == 1 {
			fmt.Printf("Mining Pending TXs with attempt %d", attempt)
		}
		fmt.Println("nonce: == ", nonce)
		block = database.NewBlock(p.parent, p.number, nonce, p.txs, p.miner)
		blockHash, err := block.Hash()
		if err != nil {
			fmt.Printf("block hash is not valid %v", err)
			return database.Block{}, err
		}
		hash = blockHash
	}
	fmt.Printf("\nMined new Block '%x' using PoW%s:\n", hash, hash)
	fmt.Printf("\tHeight: '%v'\n", block.Header.Number)
	fmt.Printf("\tNonce: '%v'\n", block.Header.Nonce)
	fmt.Printf("\tCreated: '%v'\n", block.Header.Time)
	// fmt.Printf("\tMiner: '%v'\n", block.Header.Miner)
	// fmt.Printf("\tParent: '%v'\n\n", block.Header.Parent.Hex())
	fmt.Printf("\tAttempt: '%v'\n", attempt)
	fmt.Printf("\tTime: %s\n\n", time.Since(startTime))
	return block, nil
}

func generateNonce() uint32 {
	s := rand.NewSource(time.Now().UTC().UnixNano())
	r := rand.New(s)
	return r.Uint32()
}
