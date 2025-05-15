package node

import (
	"context"
	"errors"
	"fmt"
	"taraskrasiuk/blockchain_l/internal/database"
	"time"
)

const (
	SYNC_TIME_TIMEOUT = 30 * time.Second
	DefaultHTTPport   = 8080
)

type Node struct {
	dirname        string
	port           uint
	state          *database.State
	knownPeers     map[string]PeerNode
	hasGenesisFile bool
}

func NewNode(datadir string, port uint, bootstrap *PeerNode, hasGenesisFile bool) *Node {
	node := &Node{
		dirname:        datadir,
		port:           port,
		knownPeers:     make(map[string]PeerNode),
		hasGenesisFile: hasGenesisFile,
	}

	if bootstrap != nil {
		node.knownPeers[bootstrap.TcpAddress()] = *bootstrap
	}
	return node
}

func (n *Node) Run(ctx context.Context) error {
	logger.Printf(".run() running node on port %d\n", n.port)
	// create a new state
	state, err := database.NewState(n.dirname, n.hasGenesisFile)
	if err != nil {
		return err
	}
	// assign a state to node
	n.state = state

	// run sync
	go n.sync(ctx)

	return nil
}

func (n *Node) Close() error {
	return n.state.Close()
}

func (n *Node) sync(ctx context.Context) {
	t := time.NewTicker(SYNC_TIME_TIMEOUT)
	for {
		select {
		case <-t.C:
			logger.Println(".sync(), time to sync a node")
			n.doSync(ctx)
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

func (n *Node) doSync(ctx context.Context) {
	ctxWithTimout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for _, peer := range n.knownPeers {
		status, err := peer.getPeerNodeStatus(ctxWithTimout)
		if err != nil {
			logger.Printf(".doSync() queryNodeStatus error occured %v\n", err)
		}
		err = n.syncBlocks(ctx, peer, status)
		if err != nil {
			logger.Printf(".doSync() syncBlocks error occured %v\n", err)
		}
		err = n.syncPeers(status)
		if err != nil {
			logger.Printf(".doSync() syncPeers error occured %v\n", err)
		}
	}
}

func (n *Node) IsKnownPeer(p *PeerNode) bool {
	_, isKnownPeer := n.knownPeers[p.TcpAddress()]
	return isKnownPeer
}

func (n *Node) AddPeer(p *PeerNode) {
	n.knownPeers[p.TcpAddress()] = *p
}

func (n *Node) syncPeers(status GetPeerNodeStatusResponse) error {
	for _, statusPeer := range status.KnownPeers {
		if !n.IsKnownPeer(&statusPeer) {
			logger.Printf(" found new peer node %s\n", statusPeer.TcpAddress())
			n.AddPeer(&statusPeer)
		}
	}
	return nil
}

func (n *Node) syncBlocks(ctx context.Context, p PeerNode, status GetPeerNodeStatusResponse) error {
	localBlockNumber := n.state.GetLastBlock().Header.Number
	if localBlockNumber < status.BlockNumber {
		newBlocks, err := p.getNodeBlocks(ctx, *n.state.GetLastHash())
		if err != nil {
			return fmt.Errorf("%s could not retrieve the peer node's blocks. \n %v", logger.Prefix(), err)
		}
		for _, newBlock := range newBlocks.Blocks {
			n.state.AddBlock(newBlock)
		}
		logger.Printf(" done syncing blocks for peer node %s\n", p.TcpAddress())
	}
	return nil
}

// ==== node views
type NodeBalancesListRes struct {
	Hash    *database.Hash            `json:"hash"`
	Balance map[database.Account]uint `json:"balances"`
}

func (n *Node) ViewBalancesList() NodeBalancesListRes {
	return NodeBalancesListRes{
		Hash:    n.state.GetLastHash(),
		Balance: *&n.state.Balances,
	}
}

type NodeStatusRes struct {
	BlockHash   string              `json:"block_hash"`
	BlockNumber uint64              `json:"block_number"`
	KnownPeers  map[string]PeerNode `json:"known_peers"`
}

func (n *Node) ViewNodeStatus() NodeStatusRes {
	return NodeStatusRes{
		BlockHash:   n.state.GetLastHash().String(),
		BlockNumber: n.state.GetLastBlock().Header.Number,
		KnownPeers:  n.knownPeers,
	}
}

type SyncBlocksRes struct {
	Blocks []database.Block `json:"blocks"`
}

func (n *Node) ViewSyncBlocks(afterHash database.Hash) (SyncBlocksRes, error) {
	blocks, err := n.state.GetBlocksAfter(afterHash, n.dirname)
	if err != nil {
		return SyncBlocksRes{}, err
	}
	return SyncBlocksRes{blocks}, nil
}

// ==== Add new transaction
func (n *Node) AddTransaction(from, to, data string, value uint) (database.Hash, error) {
	if from == "" || to == "" {
		return database.Hash{}, errors.New("the fields 'from' or 'to' are missed")
	}
	if value == 0 {
		return database.Hash{}, errors.New("the value could not being negative")
	}

	newTx := database.NewTx(database.Account(from), database.Account(to), data, value)
	if err := n.state.Add(*newTx); err != nil {
		return database.Hash{}, err
	}
	hash, err := n.state.Persist()
	if err != nil {
		return database.Hash{}, err
	}
	return hash, nil
}
