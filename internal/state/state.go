package state

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"taraskrasiuk/blockchain_l/internal/block"
	"taraskrasiuk/blockchain_l/internal/transactions"
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

	blockFile     *os.File
	LastBlockHash block.Hash
}

func (s *State) Close() {
	defer s.blockFile.Close()
}

func NewState(dirname string) (*State, error) {
	s := State{
		Balances: make(map[transactions.Account]uint),
	}

	if err := initDbDirStructureIfNotExist(dirname); err != nil {
		return nil, err
	}

	if err := s.loadGenesisFile(dirname); err != nil {
		return nil, err
	}
	if err := s.loadBlocksFile(dirname); err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *State) loadBlocksFile(dirname string) error {
	f, err := os.OpenFile(getBlocksDbFile(dirname), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	s.blockFile = f

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		var blockFS block.BlockFS
		if err := json.Unmarshal(scanner.Bytes(), &blockFS); err != nil {
			return err
		}
		for _, tx := range blockFS.Value.Payload {
			if err := s.apply(tx); err != nil {
				return err
			}
		}
		s.LastBlockHash = blockFS.Key
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}

// Load genesis file
func (s *State) loadGenesisFile(dirname string) error {
	f, err := os.OpenFile(getGenesisFile(dirname), os.O_RDONLY, 0600)
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

func (s *State) Add(tx transactions.Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}
	s.txMempool = append(s.txMempool, tx)
	return nil
}

// Add block to state, and apply all block's transactions to the current state txMempool.
func (s *State) AddBlock(b block.Block) error {
	for _, tx := range b.Payload {
		if err := s.apply(tx); err != nil {
			return err
		}
		s.txMempool = append(s.txMempool, tx)
	}
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

func (s *State) Persist() (block.Hash, error) {
	// create a new block, and set a parent block's hash
	b := block.NewBlock(s.LastBlockHash, s.txMempool)
	bhash, err := b.Hash()
	if err != nil {
		return block.Hash{}, err
	}
	blockFs := &block.BlockFS{
		Key:   bhash,
		Value: b,
	}
	// encode block to json
	jsonBlock, err := json.Marshal(blockFs)
	if err != nil {
		return block.Hash{}, err
	}

	if _, err := s.blockFile.Write(append(jsonBlock, '\n')); err != nil {
		return block.Hash{}, err
	}

	s.LastBlockHash = bhash

	return bhash, nil
}

func (s *State) GetVersion() string {
	return hex.EncodeToString(s.LastBlockHash[:])
}
