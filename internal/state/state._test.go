package state

import (
	"context"
	"fmt"
	"log"
	"os"
	"taraskrasiuk/blockchain_l/internal/block"
	"taraskrasiuk/blockchain_l/internal/transactions"
	"testing"
)

func f() (cancel func()) {
	_, cancel = context.WithCancel(context.TODO())
	cancel2 := func() {
		cancel()
		fmt.Println("canceled")
	}
	return cancel2
}

var testDbDir = "test-dir"

// change db files which state uses.
func setupMockGenesisDBFile(filename string) error {
	content := `{
			"genesis_time": "2019-03-18T00:00:00.000000000Z",
		    "chain_id": "the-blockchain-bar-ledger",
		    "balances": {
		        "andrej": 1000000
		    }
		}`
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(content))
	if err != nil {
		return err
	}
	return nil
}

func setupMockTxDBFile(filename string) error {
	content := `{"hash":"96a7306a8be62774de3799f1e605026327203d7eddf979ddee344eeeca05673a","block":{"header":{"ParentHash":"0000000000000000000000000000000000000000000000000000000000000000","Time":1746709322},"payload":[{"from":"andrej","to":"andrej","value":3,"data":"","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"andrej","to":"andrej","value":700,"data":"reward","createdAt":"2025-05-08T16:02:02+03:00"}]}}
{"hash":"35aa4cb8eb3c14f56563a85868339eef9de62835e1bc80827f97a471abb5db3d","block":{"header":{"ParentHash":"96a7306a8be62774de3799f1e605026327203d7eddf979ddee344eeeca05673a","Time":1746709322},"payload":[{"from":"andrej","to":"andrej","value":3,"data":"","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"andrej","to":"andrej","value":700,"data":"reward","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"andrej","to":"babayaga","value":2000,"data":"","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"andrej","to":"andrej","value":100,"data":"reward","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"babayaga","to":"andrej","value":1,"data":"","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"babayaga","to":"caesar","value":1000,"data":"","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"babayaga","to":"andrej","value":50,"data":"","createdAt":"2025-05-08T16:02:02+03:00"},{"from":"andrej","to":"andrej","value":600,"data":"reward","createdAt":"2025-05-08T16:02:02+03:00"}]}}`
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(content))
	if err != nil {
		return err
	}
	return nil
}

func setup() error {
	// change original filenames which state uses
	genesisFile = "utest_" + genesisFile
	blocksFile = "utest_" + blocksFile

	s, _ := NewState(testDbDir)
	block0 := block.NewBlock(block.Hash{}, 1, []transactions.Tx{
		*transactions.NewTx(transactions.Account("andrej"), transactions.Account("andrej"), "", 3),
		*transactions.NewTx(transactions.Account("andrej"), transactions.Account("andrej"), "reward", 700),
	})
	s.AddBlock(block0)
	block0Hash, err := s.Persist()
	if err != nil {
		log.Fatal(err)
	}
	block1 := block.NewBlock(block0Hash, 2, []transactions.Tx{
		*transactions.NewTx("andrej", "babayaga", "", 2000),
		*transactions.NewTx("andrej", "andrej", "reward", 100),
		*transactions.NewTx("babayaga", "andrej", "", 1),
		*transactions.NewTx("babayaga", "caesar", "", 1000),
		*transactions.NewTx("babayaga", "andrej", "", 50),
		*transactions.NewTx("andrej", "andrej", "reward", 600),
	})
	s.AddBlock(block1)
	s.Persist()
	return nil
}

func cleanup() error {
	// if err := os.Remove(genesisFile); err != nil {
	// 	return err
	// }
	// if err := os.Remove(blocksFile); err != nil {
	// 	return err
	// }
	if err := os.RemoveAll(testDbDir); err != nil {
		return err
	}
	return nil
}

func TestState(t *testing.T) {
	err := setup()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			log.Fatal(err)
		}
	}()

	state, _ := NewState(testDbDir)

	expectedBalance := 1000151
	if state.Balances["andrej"] != uint(expectedBalance) {
		t.Fatalf("expected the balance for andrej to be %d but got %d", expectedBalance, state.Balances["andrej"])
	}

	expectedBalance = 949
	if state.Balances["babayaga"] != uint(expectedBalance) {
		t.Fatalf("expected the balance for babayaga to be %d but got %d", expectedBalance, state.Balances["babayaga"])
	}
}
