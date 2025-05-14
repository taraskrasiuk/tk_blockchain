package node

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"taraskrasiuk/blockchain_l/internal/block"
	"taraskrasiuk/blockchain_l/internal/state"
	"taraskrasiuk/blockchain_l/internal/transactions"
	"time"
)

// PeerNode
type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint   `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

func (p *PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}

type Node struct {
	dirname        string
	port           uint
	state          *state.State
	knownPeers     map[string]PeerNode
	hasGenesisFile bool
}

type BalancesListResponse struct {
	Hash     *block.Hash                   `json:"hash"`
	Balances map[transactions.Account]uint `json:"balances"`
}

type HttpNodeHandler struct {
	n *Node
}

func NewHttpNodeHanlder(n *Node) *HttpNodeHandler {
	return &HttpNodeHandler{n}
}

func (h *HttpNodeHandler) handlerBalancesList(w http.ResponseWriter, r *http.Request) {
	balancesListResponse := BalancesListResponse{
		Hash:     h.n.state.GetLastHash(),
		Balances: h.n.state.Balances,
	}

	if err := writeJSON(w, http.StatusOK, balancesListResponse); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}
}

type NodeStatusResponse struct {
	BlockHash   string              `json:"block_hash"`
	BlockNumber uint64              `json:"block_number"`
	KnownPeers  map[string]PeerNode `json:"known_peers"`
}

func (h *HttpNodeHandler) handlerNodeStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	nodeResp := &NodeStatusResponse{
		BlockHash:   hex.EncodeToString(h.n.state.GetLastHash()[:]),
		BlockNumber: h.n.state.GetLastBlock().Header.Number,
		KnownPeers:  h.n.knownPeers,
	}
	if err := writeJSON(w, http.StatusOK, nodeResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}
}

type SyncResponse struct {
	Blocks []block.Block `json:"blocks"`
}

func (h *HttpNodeHandler) handlerSync(w http.ResponseWriter, r *http.Request) {
	reqHash := r.URL.Query().Get("fromBlock")
	if reqHash == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
		return
	}
	hash := block.Hash{}
	err := hash.UnmarshalText([]byte(reqHash))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("could not read a hash"))
		return
	}
	blocks, err := state.GetBlocksAfter(hash, h.n.dirname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not get the blocks"))
		return
	}
	writeJSON(w, http.StatusOK, SyncResponse{blocks})
}

type TxAddRequestBody struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Data  string `json:"data"`
	Value uint   `json:"value"`
}

type TxAddResponse struct {
	Hash *block.Hash `json:"hash"`
}

// TODO: move validation to transaction, and make it in NewTx
func validateTxAddReqBody(txReqBody TxAddRequestBody) error {
	if txReqBody.From == "" || txReqBody.To == "" {
		return errors.New("the fields 'from' or 'to' are missed")
	}
	if txReqBody.Value == 0 {
		return errors.New("the value could not being negative")
	}
	return nil
}

func (h *HttpNodeHandler) handlerTxAddRequest(w http.ResponseWriter, r *http.Request) {
	var txReqBody TxAddRequestBody

	if err := json.NewDecoder(r.Body).Decode(&txReqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
	defer r.Body.Close()

	if err := validateTxAddReqBody(txReqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	newTx := transactions.NewTx(transactions.NewAccount(txReqBody.From), transactions.Account(txReqBody.To), txReqBody.Data, txReqBody.Value)
	if err := h.n.state.Add(*newTx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	hash, err := h.n.state.Persist()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	resp := TxAddResponse{
		Hash: &hash,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) error {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return err
	}
	return nil
}

// Node constructor
func NewNode(datadir string, port uint, bootstrap *PeerNode, hasGenesisFile bool) *Node {
	if bootstrap != nil {
		knownPeers := make(map[string]PeerNode)
		knownPeers[bootstrap.TcpAddress()] = *bootstrap

		return &Node{
			dirname:        datadir,
			port:           port,
			knownPeers:     knownPeers,
			hasGenesisFile: hasGenesisFile,
		}
	}
	return &Node{
		dirname:        datadir,
		port:           port,
		knownPeers:     make(map[string]PeerNode),
		hasGenesisFile: hasGenesisFile,
	}
}

// PeerNode constructor
func NewPeerNode(ip string, port uint, isBootstrap, isActive bool) *PeerNode {
	return &PeerNode{ip, port, isBootstrap, isActive}
}

const DefaultHTTPport = 8080

func (n *Node) Run(ctx context.Context) error {
	fmt.Println("A node is running on port 8080")
	// create new state
	s, err := state.NewState(n.dirname, n.hasGenesisFile)
	if err != nil {
		return err
	}
	defer s.Close()
	// assign a newly state to a node state as ref
	n.state = s

	// run sync
	go n.sync(ctx)

	// run the http server
	mux := http.NewServeMux()
	nodeHandler := NewHttpNodeHanlder(n)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("status: OK"))
	})

	mux.HandleFunc("GET /balances/list", nodeHandler.handlerBalancesList)
	mux.HandleFunc("POST /tx/add", nodeHandler.handlerTxAddRequest)
	mux.HandleFunc("GET /node/status", nodeHandler.handlerNodeStatus)
	mux.HandleFunc("GET /node/sync", nodeHandler.handlerSync)

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), mux)
}

// Sync logic in order to synchronized the db of the nodes.
func (n *Node) sync(ctx context.Context) {
	t := time.NewTicker(30 * time.Second)
	// run infinite loop
	for {
		select {
		case <-t.C:
			fmt.Println("Sync time. Searching for a new blocks")

			n.doSync()
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

func (n *Node) doSync() {
	for _, peer := range n.knownPeers {
		status, err := queryNodeStatus(&peer)
		if err != nil {
			fmt.Printf("doSync() queryNodeStatus error occured %v\n", err)
		}
		err = n.syncBlocks(peer, status)
		if err != nil {
			fmt.Printf("doSync() syncBlocks error occured %v\n", err)
		}
		err = n.syncPeers(status)
		if err != nil {
			fmt.Printf("doSync() syncPeers error occured %v\n", err)
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

func (n *Node) syncPeers(status NodeStatusResponse) error {
	for _, statusPeer := range status.KnownPeers {
		if !n.IsKnownPeer(&statusPeer) {
			fmt.Printf("Found new Peer %s\n", statusPeer.TcpAddress())
			n.AddPeer(&statusPeer)
		}
	}
	return nil
}

func (n *Node) syncBlocks(p PeerNode, status NodeStatusResponse) error {
	localBlockNumber := n.state.GetLastBlock().Header.Number
	if localBlockNumber < status.BlockNumber {
		// _ := status.BlockNumber - localBlockNumber
		newBlocks, err := queryNodeBlocks(*n.state.GetLastHash(), &p)
		if err != nil {
			return fmt.Errorf("got an error while trying to get the new blocks, %v\n", err)
		}
		for _, newBlock := range newBlocks.Blocks {
			n.state.AddBlock(newBlock)
		}
		fmt.Println("done syncing blocks")
	}
	return nil
}

// HTTP request in order to get the peer node status
func queryNodeStatus(p *PeerNode) (NodeStatusResponse, error) {
	// TODO: remove the http
	url := fmt.Sprintf("http://%s/node/status", p.TcpAddress())
	fmt.Printf("queryNodeStatus() for %s", url)
	response, err := http.Get(url)
	if err != nil {
		return NodeStatusResponse{}, err
	}
	defer response.Body.Close()

	var statusResp NodeStatusResponse
	if err := json.NewDecoder(response.Body).Decode(&statusResp); err != nil {
		return NodeStatusResponse{}, err
	}
	jsonStatus, _ := json.MarshalIndent(statusResp, "", "  ")
	fmt.Printf("Got a status from nodepeer tcp: %s, and a status:\n %s\n", p.TcpAddress(), jsonStatus)
	return statusResp, nil
}

func queryNodeBlocks(lastBlockHash block.Hash, p *PeerNode) (SyncResponse, error) {
	fmt.Println("last hash: " + lastBlockHash.ToString())
	response, err := http.Get(fmt.Sprintf("http://%s/node/sync?fromBlock=%s", p.TcpAddress(), lastBlockHash.ToString()))
	if err != nil {
		return SyncResponse{}, err
	}
	defer response.Body.Close()
	var statusResp SyncResponse
	if err := json.NewDecoder(response.Body).Decode(&statusResp); err != nil {
		return SyncResponse{}, err
	}
	return statusResp, nil
}
