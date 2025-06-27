package node

import (
	"context"
	"fmt"
	"os"
	"sync"
	"taraskrasiuk/blockchain_l/internal/database"
	"testing"
	"time"
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

func setup() {
	if err := os.Mkdir(testDir, 0777); err != nil {
		panic(err)
	}
	if err := setupMockGenesisDBFile(testDir + "/genesis.json"); err != nil {
		panic(err)
	}

}

func clear() {
	if err := os.RemoveAll(testDir); err != nil {
		panic(err)
	}
}

var testDir = "ttest"

func TestNode_Run(t *testing.T) {
	setup()
	defer clear()

	pctx := context.Background()
	ctx, cancel := context.WithTimeout(pctx, 5*time.Second)
	defer cancel()

	miner := database.NewAccount("miner")
	n := NewNode(testDir, 8080, "", nil, miner, true)
	if err := n.Run(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestNode_Mining(t *testing.T) {
	setup()
	defer clear()

	miner := database.NewAccount("miner")
	peerNode := NewPeerNode("localhost", 8080, true, true)
	n := NewNode(testDir, 8081, "localhost", peerNode, miner, true)
	pctx := context.Background()
	ctx, cancel := context.WithTimeout(pctx, 60*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)
	// create 1 transaction
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		tx := database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("taras"), "", 100), []byte{})

		if err := n.AddPendingTX(*tx); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(6 * time.Second)
		tx := database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("taras"), "", 300), []byte{})

		if err := n.AddPendingTX(*tx); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				if n.state.GetLastBlock().Header.Number == 1 {
					fmt.Println("height : ", n.state.GetLastBlock().Header.Number)
					if err := n.Close(); err != nil {
						panic(err)
					}
					return
				}
			}
		}
	}()

	if err := n.Run(ctx); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if n.state.GetLastBlock().Header.Number > 1 {
		t.Fatalf("expected a block heaight to be 1 but got %d", n.state.GetLastBlock().Header.Number)
	}
}

func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	MINE_PENDING_INTERVAL = 10 * time.Second
	setup()
	defer clear()
	var wg sync.WaitGroup

	tarasAcc := database.NewAccount("taras")
	p := NewPeerNode("localhost", 8080, true, false)
	n := NewNode(testDir, 8081, "localhost", p, tarasAcc, true)

	tx1 := database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("taras"), "", 100), []byte{})
	tx2 := database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("taras"), "", 5), []byte{})
	tx2Hash, _ := tx2.Hash()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	firstPendingBlock := NewPendingBlock(database.Hash{}, 0, []database.SignedTx{*tx1}, tarasAcc)
	validSyncedBlock, err := Mine(ctx, firstPendingBlock)
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(3)
	go func() {
		defer wg.Done()
		time.Sleep(MINE_PENDING_INTERVAL - (MINE_PENDING_INTERVAL / 4))
		err := n.AddPendingTX(*tx1)
		if err != nil {
			panic(err)
		}
		err = n.AddPendingTX(*tx2)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(MINE_PENDING_INTERVAL + (2 * time.Second))
		if !n.isMining {
			panic("should be mining")
		}
		_, err := n.state.AddBlock(validSyncedBlock)
		if err != nil {
			panic(err)
		}
		n.newSyncedBlocksCh <- validSyncedBlock

		time.Sleep(time.Second * 2)
		if n.isMining {
			t.Fatal("synced block should have canceled mining")
		}
		// Mined TX1 by Andrej should be removed from the Mempool
		_, onlyTX2IsPending := n.pendingTXs[tx2Hash.String()]
		if len(n.pendingTXs) != 1 && !onlyTX2IsPending {
			t.Fatal("TX1 should be still pending")
		}
		time.Sleep(MINE_PENDING_INTERVAL * 2)
		if !n.isMining {
			t.Fatal("should attempt to mine TX1 again")
		}
	}()
	go func() {
		defer wg.Done()
		// Regularly check whenever both TXs are now mined
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				if n.state.GetLastBlock().Header.Number == 1 {
					if err := n.Close(); err != nil {
						panic(err)
					}
					return
				}
			}
		}
	}()

	if err := n.Run(ctx); err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	if n.state.GetLastBlock().Header.Number > 1 {
		t.Fatalf("expected a block heaight to be 1 but got %d", n.state.GetLastBlock().Header.Number)
	}
	// check miner balance
	balance := n.state.Balances[tarasAcc]
	expectedBalance := 105 + 175 + 175
	if balance != uint(expectedBalance) /* with 2 rewards */ {
		t.Fatalf("expected balance for miner account to be %d but got %d", expectedBalance, balance)
	}
}
