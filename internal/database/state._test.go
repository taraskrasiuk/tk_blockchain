package database

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

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
	acc := common.HexToAddress("miner")

	s, _ := NewState(testDbDir, true)
	block0 := NewBlock(Hash{}, 1, 0x0123, []SignedTx{
		*NewSignedTx(*NewTx(NewAccount("andrej"), NewAccount("andrej"), "", 3, s.NextAccountNonce(acc)), []byte{}),
		*NewSignedTx(*NewTx(NewAccount("andrej"), NewAccount("andrej"), "reward", 700, s.NextAccountNonce(acc)), []byte{}),
	}, NewAccount("miner"))
	s.AddBlock(block0)
	block0Hash, err := s.AddBlock(block0)
	if err != nil {
		log.Fatal(err)
	}
	block1 := NewBlock(block0Hash, 2, 0x0123, []SignedTx{
		*NewSignedTx(*NewTx(NewAccount("andrej"), NewAccount("babayaga"), "", 2000, s.NextAccountNonce(acc)), []byte{}),
		*NewSignedTx(*NewTx(NewAccount("andrej"), NewAccount("andrej"), "reward", 100, s.NextAccountNonce(acc)), []byte{}),
		*NewSignedTx(*NewTx(NewAccount("babayaga"), NewAccount("andrej"), "", 1, s.NextAccountNonce(acc)), []byte{}),
		*NewSignedTx(*NewTx(NewAccount("babayaga"), NewAccount("caesar"), "", 1000, s.NextAccountNonce(acc)), []byte{}),
		*NewSignedTx(*NewTx(NewAccount("babayaga"), NewAccount("andrej"), "", 50, s.NextAccountNonce(acc)), []byte{}),
		*NewSignedTx(*NewTx(NewAccount("andrej"), NewAccount("andrej"), "reward", 600, s.NextAccountNonce(acc)), []byte{}),
	}, NewAccount("miner"))
	s.AddBlock(block1)
	return nil
}

func cleanup() error {
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

	state, _ := NewState(testDbDir, true)

	expectedBalance := 999451
	if state.Balances[NewAccount("andrej")] != uint(expectedBalance) {
		t.Fatalf("expected the balance for andrej to be %d but got %d", expectedBalance, state.Balances[NewAccount("andrej")])
	}

	expectedBalance = 949
	if state.Balances[NewAccount("babayaga")] != uint(expectedBalance) {
		t.Fatalf("expected the balance for babayaga to be %d but got %d", expectedBalance, state.Balances[NewAccount("babayaga")])
	}
}

func Test_Acc(t *testing.T) {
	a := NewAccount("qwe")
	fmt.Println(a)
}
