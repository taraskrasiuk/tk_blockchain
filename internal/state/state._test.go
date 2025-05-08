package state

import (
	"log"
	"os"
	"testing"
)

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
	content := `{"from":"andrej","to":"andrej","value":3,"data":""}
{"from":"andrej","to":"andrej","value":700,"data":"reward"}
{"from":"andrej","to":"babayaga","value":2000,"data":""}
{"from":"andrej","to":"andrej","value":100,"data":"reward"}
{"from":"babayaga","to":"andrej","value":1,"data":""}`
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
	blocksFile = "utest_" +blocksFile 

	err := setupMockGenesisDBFile(genesisFile)
	if err != nil {
		return err
	}
	err = setupMockTxDBFile(blocksFile)
	if err != nil {
		return err
	}
	return nil
}

func cleanup() error {
	err := os.Remove(genesisFile)
	if err != nil {
		return err
	}
	err = os.Remove(blocksFile)
	if err != nil {
		return err
	}
	return nil
}

func TestState(t *testing.T) {
	err := setup()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	state := NewState("")

	expectedBalance := 998801
	if state.Balances["andrej"] != uint(expectedBalance) {
		t.Fatalf("expected the balance for andrej to be %d but got %d", expectedBalance, state.Balances["andrej"])
	}

	expectedBalance = 1999
	if state.Balances["babayaga"] != uint(expectedBalance) {
		t.Fatalf("expected the balance for babayaga to be %d but got %d", expectedBalance, state.Balances["babayaga"])
	}
}
