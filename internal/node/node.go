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
	dirname    string
	port       uint
	state      *state.State
	knownPeers map[string]PeerNode
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
func NewNode(datadir string, port uint, bootstrap *PeerNode) *Node {
	if bootstrap != nil {
		knownPeers := make(map[string]PeerNode)
		knownPeers[bootstrap.TcpAddress()] = *bootstrap

		return &Node{
			dirname:    datadir,
			port:       port,
			knownPeers: knownPeers,
		}
	}
	return &Node{
		dirname:    datadir,
		port:       port,
		knownPeers: make(map[string]PeerNode),
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
	s, err := state.NewState(n.dirname)
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

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), mux)
}

// Sync logic in order to synchronized the db of the nodes.
func (n *Node) sync(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)

	// run infinite loop
	for {
		select {
		case <-t.C:
			fmt.Println("Sync time. Searching for a new blocks")

			n.lookupNewBlocksAndPeers()
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

func (n *Node) lookupNewBlocksAndPeers() error {
	for _, peer := range n.knownPeers {
		status, err := queryNodeStatus(&peer)
		if err != nil {
			fmt.Errorf("got an error while looking up for a new blocks, %v\n", err)
			continue
		}
		jsonStatus, _ := json.MarshalIndent(status, "", "  ")
		fmt.Printf("Got a status from nodepeer tcp: %s, and a status:\n %s\n", peer.TcpAddress(), jsonStatus)
		localBlockNumber := n.state.GetLastBlock().Header.Number
		if localBlockNumber < status.BlockNumber {
			// _ := status.BlockNumber - localBlockNumber
		}

		for _, statusPeer := range status.KnownPeers {
			newPeer, isKnowPeer := n.knownPeers[statusPeer.TcpAddress()]
			if !isKnowPeer {
				fmt.Printf("Found a new peer %s\n", newPeer.IP)

				n.knownPeers[newPeer.TcpAddress()] = newPeer
			}
		}
	}

	return nil
}

// HTTP request in order to get the peer node status
func queryNodeStatus(p *PeerNode) (NodeStatusResponse, error) {
	// TODO: remove the http
	response, err := http.Get(fmt.Sprintf("http://%s/node/status", p.TcpAddress()))
	if err != nil {
		return NodeStatusResponse{}, err
	}
	defer response.Body.Close()
	var statusResp NodeStatusResponse
	if err := json.NewDecoder(response.Body).Decode(&statusResp); err != nil {
		return NodeStatusResponse{}, err
	}
	return statusResp, nil
}
