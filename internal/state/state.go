package state

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"taraskrasiuk/blockchain_l/internal/transactions"
)

var (
	genesisFileDb      = "genesis.db"
	transactionsFileDb = "tx.db"
)

type genesisResource struct {
	GenesisTime string `json:"genesis_time"`
	ChainID     string `json:"chain_id"`
	Balances    map[string]uint
}

type State struct {
	// all balances
	Balances  map[transactions.Account]uint
	txMempool []transactions.Tx

	dbFile   *os.File
	snapshot [32]byte
}

func (s *State) Close() {
	defer s.dbFile.Close()
}

func NewState() *State {
	s := &State{
		Balances: make(map[transactions.Account]uint),
	}

	s.loadGenesisFile()
	s.loadTransactions()
	return s
}

// Load genesis file
func (s *State) loadGenesisFile() error {
	f, err := os.OpenFile(genesisFileDb, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 128)
	res := []byte{}
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		res = append(res, buf[:n]...)
	}
	var genesisData genesisResource
	err = json.Unmarshal(res, &genesisData)
	if err != nil {
		return err
	}
	// set a balances to state
	for k, v := range genesisData.Balances {
		s.Balances[transactions.Account(k)] = v
	}

	return nil
}

// load transactions file
func (s *State) loadTransactions() error {
	f, err := os.OpenFile(transactionsFileDb, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	// dont close the file
	// defer f.Close()

	// save file ref to state
	s.dbFile = f

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var tx transactions.Tx
		err := json.Unmarshal(scanner.Bytes(), &tx)
		if err != nil {
			return err
		}

		// apply transaction
		if err := s.apply(tx); err != nil {
			return err
		}
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}

func (s *State) Add(tx transactions.Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}
	s.txMempool = append(s.txMempool, tx)
	return nil
}

func (s *State) apply(tx transactions.Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return errors.New("cannot perform transaction due to insufficient balance")
	}
	s.Balances[tx.From] -= tx.Value

	if _, ok := s.Balances[tx.To]; !ok {
		s.Balances[tx.To] = 0
	}
	s.Balances[tx.To] += tx.Value

	return nil
}

// Persistent
func (s *State) Persist() (Snapshot, error) {
	// copy mempool
	mempool := make([]transactions.Tx, len(s.txMempool))
	copy(mempool, s.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJson, err := json.Marshal(mempool[i])
		if err != nil {
			return Snapshot{}, err
		}

		if _, err := s.dbFile.Write(append(txJson, '\n')); err != nil {
			fmt.Println(err)
			return Snapshot{}, err
		}
		// create a snaphost sha256.Sum256
		err = s.doSnapshot()
		if err != nil {
			return Snapshot{}, err
		}

		s.txMempool = s.txMempool[1:]
	}

	return s.snapshot, nil
}

// Snapshot
type Snapshot [32]byte

func (s *State) doSnapshot() error {
	_, err := s.dbFile.Seek(0, 0)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(s.dbFile)
	if err != nil {
		return err
	}

	s.snapshot = sha256.Sum256(data)
	return nil
}

func (s *State) GetVersion() string {
	fmt.Println(s.snapshot)
	return fmt.Sprintf("%x", s.snapshot)
}
