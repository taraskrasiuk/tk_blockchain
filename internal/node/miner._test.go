package node

import (
	"context"
	"taraskrasiuk/blockchain_l/internal/database"
	"testing"
	"time"
)

func createRandomPendingBlock() *PendingBlock {
	minerAcc := database.NewAccount("test")
	return NewPendingBlock(database.Hash{}, 0, []database.SignedTx{
		*database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("taras"), "", 3, 1), []byte{}),
	}, minerAcc)
}

func TestMine(t *testing.T) {
	pendingBlock := createRandomPendingBlock()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		t.Fatal(err)
	}
	minedHash, err := minedBlock.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if !database.IsValidBlock(minedHash) {
		t.Fatal("the block's hash is not valid")
	}
}
