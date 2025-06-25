package node

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"taraskrasiuk/blockchain_l/internal/database"
	"time"
)

const (
	SYNC_TIME_TIMEOUT              = 10 * time.Second
	MINE_PENDING_INTERVAL_DURATION = 20 * time.Second
	DefaultHTTPport                = 8080
)

// Node Config

type Node struct {
	dirname        string
	ip             string
	port           uint
	state          *database.State
	mu             sync.Mutex
	knownPeers     map[string]PeerNode
	hasGenesisFile bool // TODO: probably no need
	// mining
	pendingTXs        map[string]database.Tx
	archivedTXs       map[string]database.Tx
	newSyncedBlocksCh chan database.Block
	newPendingTXsCh   chan database.Tx
	isMining          bool
	// public
	IsBootstrap bool
}

func NewNode(datadir string, port uint, ip string, bootstrap *PeerNode, hasGenesisFile bool) *Node {
	node := &Node{
		dirname:           datadir,
		ip:                ip,
		port:              port,
		knownPeers:        make(map[string]PeerNode),
		hasGenesisFile:    hasGenesisFile,
		pendingTXs:        make(map[string]database.Tx),
		archivedTXs:       make(map[string]database.Tx),
		newSyncedBlocksCh: make(chan database.Block),
		newPendingTXsCh:   make(chan database.Tx, 500),
		isMining:          false,
	}

	if bootstrap != nil {
		node.knownPeers[bootstrap.TcpAddress()] = *bootstrap
	} else {
		node.IsBootstrap = true
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
	go n.mine(ctx)

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
		if err := n.syncPendingTXs(&peer, status); err != nil {
			logger.Printf(".syncPendingTXs error occured %v\n", err)
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

func (n *Node) syncPendingTXs(p *PeerNode, status GetPeerNodeStatusResponse) error {
	for _, tx := range status.PendingTXs {
		if err := n.AddPendingTX(tx); err != nil {
			return err
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
			_, err := n.state.AddBlock(newBlock)
			if err != nil {
				return err
			}

			// Need to notify the Miner logic, in order to stop processing pending transactions,
			// due to incommed new block
			n.newSyncedBlocksCh <- newBlock
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
	PendingTXs  []database.Tx       `json:"pendingTXs"`
}

func (n *Node) ViewNodeStatus() NodeStatusRes {
	return NodeStatusRes{
		BlockHash:   n.state.GetLastHash().String(),
		BlockNumber: n.state.GetLastBlock().Header.Number,
		KnownPeers:  n.knownPeers,
		PendingTXs:  n.pendingTXsToArray(),
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

// Node mining process.
func (n *Node) mine(ctx context.Context) error {
	// The time interval
	d := MINE_PENDING_INTERVAL_DURATION
	if n.IsBootstrap {
		d = 60 * time.Second
	}
	ticker := time.NewTicker(d)
	var (
		innerContext      context.Context
		cancelContextFunc context.CancelFunc
	)
	for {
		select {
		// handle ticker case
		case <-ticker.C:
			go func() {
				if !n.isMining && len(n.pendingTXs) > 0 {
					n.isMining = true

					innerContext, cancelContextFunc = context.WithCancel(ctx)
					if err := n.processPendingTXs(innerContext); err != nil {
						// TODO what to do with an error
						fmt.Println("ERROR ", err)
					}
					n.isMining = false
				}
			}()
		case block, _ := <-n.newSyncedBlocksCh:
			// check if current node is in mining process
			if n.isMining {
				fmt.Println("Another peer node mined the new block faster, need to cancel current mining.")
				n.removeMindedPendingTXs(block)
				// cancel current mining process
				cancelContextFunc()
			}
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (n *Node) removeMindedPendingTXs(block database.Block) error {
	if len(block.Payload) > 0 && len(n.pendingTXs) > 0 {
		fmt.Println("removeMindedPendingTXs, update in memory pending transactions")
	}
	for _, tx := range block.Payload {
		txHash, err := tx.Hash()
		if err != nil {
			return err
		}
		// check the pending TXs map contains a transcation by hash
		if _, exists := n.pendingTXs[txHash.String()]; exists {
			n.archivedTXs[txHash.String()] = tx
			delete(n.pendingTXs, txHash.String())
		}
	}
	return nil
}

func (n *Node) processPendingTXs(ctx context.Context) error {
	pendingBlock := NewPendingBlock(*n.state.GetLastHash(), n.state.NextBlockNumber(), n.pendingTXsToArray())
	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		return err
	}

	if err := n.removeMindedPendingTXs(minedBlock); err != nil {
		return err
	}
	_, err = n.state.AddBlock(minedBlock)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) AddPendingTX(tx database.Tx) error {
	if err := n.state.IsValidTX(tx); err != nil {
		return err
	}
	txHash, err := tx.Hash()
	if err != nil {
		return err
	}
	_, isAlreadyPending := n.pendingTXs[txHash.String()]
	_, isArchived := n.archivedTXs[txHash.String()]
	if !isAlreadyPending && !isArchived {
		txJson, err := json.Marshal(tx)
		if err != nil {
			return err
		}
		fmt.Printf("Added Pending TX %s from Peer %s\n", txJson, n.ip)
		n.pendingTXs[txHash.String()] = tx
	}

	return nil
}

func (n *Node) pendingTXsToArray() []database.Tx {
	result := make([]database.Tx, len(n.pendingTXs))
	i := 0
	for _, tx := range n.pendingTXs {
		result[i] = tx
		i++
	}

	return result
}
