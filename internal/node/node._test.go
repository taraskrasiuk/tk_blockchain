package node

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"taraskrasiuk/blockchain_l/internal/database"
	"taraskrasiuk/blockchain_l/internal/wallet"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

var (
	testDir     = "ttest"
	passphrase1 = "qwe123QWE!@#"
	passphrase2 = "qwe123QWE!@#"
)

func setupMockAccounts(pwds ...string) ([]accounts.Account, error) {
	keystore := keystore.NewKeyStore(testDir+"/keystore", keystore.StandardScryptN, keystore.StandardScryptP)
	var res []accounts.Account
	for _, pwd := range pwds {
		acc, err := keystore.NewAccount(pwd)
		if err != nil {
			return nil, err
		}
		res = append(res, acc)
	}
	return res, nil
}

// change db files which state uses.
func setupMockGenesisDBFile(filename string) ([]accounts.Account, error) {
	accs, err := setupMockAccounts(passphrase1, passphrase2)
	if err != nil {
		return nil, err
	}
	gen := database.NewGenesisResource()
	for _, acc := range accs {
		gen.AddAccount(acc.Address.Hex(), 1000)
	}
	if err := gen.SaveToFile(filename); err != nil {
		return nil, err
	}
	return accs, nil
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

func setup() []accounts.Account {
	if err := os.Mkdir(testDir, 0777); err != nil {
		panic(err)
	}
	if err := os.Mkdir(testDir+"/database", 0777); err != nil {
		panic(err)
	}
	if accs, err := setupMockGenesisDBFile(testDir + "/database/genesis.json"); err != nil {
		panic(err)
	} else {
		return accs
	}
}

func clear() {
	if err := os.RemoveAll(testDir); err != nil {
		panic(err)
	}
}

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
	accounts := setup()
	defer clear()

	miner := accounts[0].Address
	peerNode := NewPeerNode("localhost", 8080, true, true)
	n := NewNode(testDir, 8081, "localhost", peerNode, miner, true)

	pctx := context.Background()
	ctx, cancel := context.WithTimeout(pctx, 20*time.Minute)
	defer cancel()

	var (
		acc1 = accounts[0].Address
		acc2 = accounts[1].Address
		wg   = sync.WaitGroup{}
	)

	wg.Add(3)
	// create 1 transaction
	go func() {
		defer wg.Done()
		time.Sleep(2 * time.Second)
		tx := database.NewTx(acc1, acc2, "", 100)

		signedTx, err := wallet.SignTxWithKeystoreAccount(*tx, acc1, passphrase1, wallet.GetKeystoreDirPath(n.Dirname()))
		if err != nil {
			log.Fatal(err)
		}

		if err := n.AddPendingTX(signedTx); err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(6 * time.Second)
		tx := database.NewTx(acc1, acc2, "", 300)

		signedTx, err := wallet.SignTxWithKeystoreAccount(*tx, acc1, passphrase1, wallet.GetKeystoreDirPath(n.Dirname()))
		if err != nil {
			log.Fatal(err)
		}
		if err := n.AddPendingTX(signedTx); err != nil {
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
	accounts := setup()
	defer clear()

	miner := accounts[0].Address
	peerNode := NewPeerNode("localhost", 8080, true, true)
	n := NewNode(testDir, 8081, "localhost", peerNode, miner, true)

	pctx := context.Background()
	ctx, cancel := context.WithTimeout(pctx, 20*time.Minute)
	defer cancel()

	var (
		acc1 = accounts[0].Address
		acc2 = accounts[1].Address
		wg   = sync.WaitGroup{}
	)
	tx1 := database.NewTx(acc1, acc2, "", 100)
	signedTx1, err := wallet.SignTxWithKeystoreAccount(*tx1, acc1, passphrase1, wallet.GetKeystoreDirPath(n.Dirname()))
	if err != nil {
		t.Fatal(err)
	}
	tx2 := database.NewTx(acc1, acc2, "", 200)
	signedTx2, err := wallet.SignTxWithKeystoreAccount(*tx2, acc1, passphrase1, wallet.GetKeystoreDirPath(n.Dirname()))
	if err != nil {
		t.Fatal(err)
	}
	tx2Hash, _ := tx2.Hash()

	firstPendingBlock := NewPendingBlock(database.Hash{}, 0, []database.SignedTx{signedTx1}, acc1)
	validSyncedBlock, err := Mine(ctx, firstPendingBlock)
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(3)
	go func() {
		defer wg.Done()
		time.Sleep(MINE_PENDING_INTERVAL - (MINE_PENDING_INTERVAL / 4))
		err := n.AddPendingTX(signedTx1)
		if err != nil {
			panic(err)
		}
		err = n.AddPendingTX(signedTx2)
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
	balance := n.state.Balances[acc1]
	expectedBalance := 1000 - 100 - 200 + 175 + 175
	if balance != uint(expectedBalance) /* with 2 rewards */ {
		t.Fatalf("expected balance for miner account to be %d but got %d", expectedBalance, balance)
	}
}
