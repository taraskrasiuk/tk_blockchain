package node

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"taraskrasiuk/blockchain_l/internal/database"
	"time"
)

const (
	SYNC_TIME_TIMEOUT = 15 * time.Second
	DefaultHTTPport   = 8080
)

type Node struct {
	dirname        string
	ip             string
	port           uint
	state          *database.State
	mu             sync.Mutex
	knownPeers     map[string]PeerNode
	hasGenesisFile bool
}

func NewNode(datadir string, port uint, ip string, bootstrap *PeerNode, hasGenesisFile bool) *Node {
	node := &Node{
		dirname:        datadir,
		ip:             ip,
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
		if peer.IP == n.ip && peer.Port == n.port {
			logger.Println(".doSync() skip sync self")
			continue
		}
		logger.Printf(".doSync() running for peer: %s\n", peer.TcpAddress())
		status, err := peer.getPeerNodeStatus(ctxWithTimout)
		if err != nil {
			logger.Printf(".doSync() queryNodeStatus error occured %v\n", err)
		}
		err = n.joinPeer(ctx, &peer)
		if err != nil {
			logger.Printf(".doSync() joining peer %s", peer.TcpAddress())
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

// This function will send a request to node in order to be added to known peers.
func (n *Node) joinPeer(ctx context.Context, p *PeerNode) error {
	// skip if node already connected to peer node
	if p.IsActive {
		logger.Printf(" joinPeer() peer is active")
		return nil
	}
	url := fmt.Sprintf("%s/node/addpeer?ip=%s&port=%d", p.TcpAddressWithProtocol(), n.ip, n.port)
	logger.Printf(" joinPeer() with url %s", url)
	if _, err := http.Get(url); err != nil {
		return err
	}
	response, err := p.joinPeer(ctx, n.ip, n.port)
	if err != nil {
		logger.Printf(" joinPerr() got an error %v", err)
		return err
	}
	logger.Printf(" joinPeer(), receive a response %v", response)
	if response.Error != "" {
		return fmt.Errorf(response.Error)
	}

	// update the isActive field
	n.mu.Lock()
	p.IsActive = response.Success
	n.AddPeer(p)
	n.mu.Unlock()
	logger.Printf("Successfully sending a request to add a node with ip %s to peer node %s", n.ip, p.TcpAddress())

	return nil
}

func (n *Node) syncBlocks(ctx context.Context, p PeerNode, status GetPeerNodeStatusResponse) error {
	localBlockNumber := n.state.GetLastBlock().Header.Number
	if localBlockNumber < status.BlockNumber {
		logger.Printf(" getNodeBlocks() with a last hash: %s", *n.state.GetLastHash())
		newBlocks, err := p.getNodeBlocks(ctx, *n.state.GetLastHash())
		if err != nil {
			return fmt.Errorf("%s could not retrieve the peer node's blocks. \n %v", logger.Prefix(), err)
		}
		logger.Printf("Found new blocks %d", len(newBlocks.Blocks))
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
	// validate input
	if from == "" || to == "" {
		return database.Hash{}, errors.New("the fields 'from' or 'to' are missed")
	}
	if value == 0 {
		return database.Hash{}, errors.New("the value could not being negative")
	}

	logger.Println("SHOW current txMemPool:")
	for _, t := range n.state.GetMemPool() {
		logger.Printf(". current tx in mem pool: from %s, to %s , val %d", t.From, t.To, t.Value)
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
