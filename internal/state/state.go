package state

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
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

	blockFile       *os.File
	lastBlock       block.Block
	lastBlockHash   block.Hash
	hasGenesisBlock bool
}

func (s *State) Close() {
	defer s.blockFile.Close()
}

func NewState(dirname string, hasGenesisBlock bool) (*State, error) {
	s := State{
		Balances:        make(map[transactions.Account]uint),
		hasGenesisBlock: hasGenesisBlock,
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
		applyBlock(blockFS.Value, s)
		// for _, tx := range blockFS.Value.Payload {
		// 	if err := s.apply(tx); err != nil {
		// 		return err
		// 	}
		// }
		s.lastBlock = blockFS.Value
		s.lastBlockHash = blockFS.Key
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
	if err := applyTx(tx, s); err != nil {
		return err
	}
	s.txMempool = append(s.txMempool, tx)
	return nil
}

func (s *State) AddBlock(b block.Block) (block.Hash, error) {
	// make a temporary copy of the state, in order to avoid race conditions
	pendingState := s.copy()
	// apply a block to pending state
	if err := applyBlock(b, &pendingState); err != nil {
		return block.Hash{}, err
	}
	// get a block hash
	blockHash, err := b.Hash()
	if err != nil {
		return block.Hash{}, err
	}
	// create a blockFS instance for saving it to file
	blockFS := block.BlockFS{
		Key:   blockHash,
		Value: b,
	}
	blockFSjson, err := json.Marshal(&blockFS)
	if err != nil {
		return block.Hash{}, err
	}
	fmt.Println("Persisting a new block to db file")
	if _, err := s.blockFile.Write(append(blockFSjson, '\n')); err != nil {
		return block.Hash{}, err
	}
	s.Balances = pendingState.Balances
	s.lastBlockHash = blockHash
	s.lastBlock = b

	return blockHash, nil
}

func (s *State) copy() State {
	newState := State{}
	newState.lastBlockHash = s.lastBlockHash
	newState.lastBlock = s.lastBlock
	newState.txMempool = make([]transactions.Tx, len(s.txMempool))
	newState.Balances = make(map[transactions.Account]uint)

	for acc, balance := range s.Balances {
		newState.Balances[acc] = balance
	}

	for _, tx := range s.txMempool {
		newState.txMempool = append(newState.txMempool, tx)
	}
	return newState
}

// Add block to state, and apply all block's transactions to the current state txMempool.
func applyBlock(b block.Block, s *State) error {
	nextExpectedBlockNumber := s.lastBlock.Header.Number + 1

	// validate for expected next block number. The height should be equal to state last block's number + 1.
	if s.hasGenesisBlock && b.Header.Number != nextExpectedBlockNumber {
		return fmt.Errorf("the next block number is incorrect, expected to be %d got %d", nextExpectedBlockNumber, b.Header.Number)
	}
	// validate that next block parent hash equals to state last block hash.
	if s.hasGenesisBlock && s.lastBlock.Header.Number > 0 && !reflect.DeepEqual(b.Header.ParentHash, s.lastBlockHash) {
		return fmt.Errorf("the next block parent hash is incorrect, expected to be %x got %x", s.lastBlockHash, b.Header.ParentHash)
	}
	return applyTXs(b.Payload, s)
}

func applyTXs(txs []transactions.Tx, s *State) error {
	for _, tx := range txs {
		if err := applyTx(tx, s); err != nil {
			return err
		}
		s.txMempool = append(s.txMempool, tx)
	}
	return nil
}

func applyTx(tx transactions.Tx, s *State) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if s.Balances[tx.From] < tx.Value {
		return fmt.Errorf("wrong TX, cant perform transaction. \n From: %s, To: %s, Value: %d \n", tx.From, tx.To, tx.Value)
	}
	s.Balances[tx.From] -= tx.Value

	if _, ok := s.Balances[tx.To]; !ok {
		s.Balances[tx.To] = 0
	}
	s.Balances[tx.To] += tx.Value

	return nil
}

// func (s *State) apply(tx transactions.Tx) error {
// 	if tx.IsReward() {
// 		s.Balances[tx.To] += tx.Value
// 		return nil
// 	}

// 	if s.Balances[tx.From] < tx.Value {
// 		return fmt.Errorf("wrong TX, cant perform transaction. \n From: %s, To: %s, Value: %d \n", tx.From, tx.To, tx.Value)
// 	}
// 	s.Balances[tx.From] -= tx.Value

// 	if _, ok := s.Balances[tx.To]; !ok {
// 		s.Balances[tx.To] = 0
// 	}
// 	s.Balances[tx.To] += tx.Value

// 	return nil
// }

func (s *State) Persist() (block.Hash, error) {
	// create a new block, and set a parent block's hash
	b := block.NewBlock(s.lastBlockHash, s.lastBlock.Header.Number+1, s.txMempool)
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

	s.lastBlockHash = bhash
	s.lastBlock = blockFs.Value

	return bhash, nil
}

func (s *State) GetLastHash() *block.Hash {
	return &s.lastBlockHash
}

func (s *State) GetLastBlock() *block.Block {
	return &s.lastBlock
}

func GetBlocksAfter(blockHash block.Hash, dir string) ([]block.Block, error) {
	f, err := os.OpenFile(getBlocksDbFile(dir), os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	blocks := []block.Block{}
	shouldStartAppending := false

	for scanner.Scan() {
		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
		var currentBlock block.BlockFS
		if err := json.Unmarshal(scanner.Bytes(), &currentBlock); err != nil {
			return nil, err
		}

		if shouldStartAppending {
			blocks = append(blocks, currentBlock.Value)
		}

		if currentBlock.Key == blockHash {
			shouldStartAppending = true
		}
	}

	return blocks, nil
}
