package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"taraskrasiuk/blockchain_l/internal/database"
	"time"
)

// ===========
type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint   `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

func (p *PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
}

func (p *PeerNode) TcpAddressWithProtocol() string {
	return fmt.Sprintf("http://%s", p.TcpAddress())
}

func NewPeerNode(ip string, port uint, isBootstrap, isActive bool) *PeerNode {
	return &PeerNode{ip, port, isBootstrap, isActive}
}

// ==========
type GetPeerNodeStatusResponse struct {
	BlockHash   string              `json:"block_hash"`
	BlockNumber uint64              `json:"block_number"`
	KnownPeers  map[string]PeerNode `json:"known_peers"`
}

// Get a peer node status. Used a context with a timeout 1 second.
func (p *PeerNode) getPeerNodeStatus(ctx context.Context) (GetPeerNodeStatusResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	logger.Printf(".GetPeerNodeStatus running for peer node %s", p.TcpAddress())

	var statusResp GetPeerNodeStatusResponse
	result, err := getReq(ctxWithTimeout, p, "node/status", &statusResp)
	if err != nil {
		return GetPeerNodeStatusResponse{}, err
	}
	status, ok := result.(*GetPeerNodeStatusResponse)
	if !ok {
		return GetPeerNodeStatusResponse{}, fmt.Errorf("%s type assertion failed for GetPeerNodeStatusResponse", logger.Prefix())
	}
	// TODO: temporary display this message
	jsonStatus, _ := json.MarshalIndent(statusResp, "", "  ")
	logger.Printf(".GetPeerNodeStatus() get a status from peer node with tcp: %s, and a status:\n %s\n", p.TcpAddress(), jsonStatus)
	return *status, nil
}

type GetNodeBlocksResponse struct {
	Blocks []database.Block `json:"blocks"`
}

func (p *PeerNode) getNodeBlocks(ctx context.Context, lastBlockHash database.Hash) (GetNodeBlocksResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	logger.Printf(".GetNodeBlocks() running with a last hash: %s", lastBlockHash)

	var statusResp GetNodeBlocksResponse
	result, err := getReq(ctxWithTimeout, p, fmt.Sprintf("node/sync?fromBlock=%s", lastBlockHash), &statusResp)
	if err != nil {
		return GetNodeBlocksResponse{}, err
	}
	blocks, ok := result.(*GetNodeBlocksResponse)
	if !ok {
		return GetNodeBlocksResponse{}, fmt.Errorf("%s. could not convert a response to type GetNodeBlocksResponse", logger.Prefix())
	}
	return *blocks, nil
}

type GetAddingPeerResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (p *PeerNode) joinPeer(ctx context.Context, ip string, port uint) (GetAddingPeerResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var res GetAddingPeerResponse
	result, err := getReq(ctxWithTimeout, p, fmt.Sprintf("/node/addpeer?ip=%s&port=%d`", ip, port), &res)
	fmt.Printf("result :: %v", result)
	if err != nil {
		logger.Printf("e %v\n", err)
		return GetAddingPeerResponse{false, err.Error()}, err
	}
	response, ok := result.(*GetAddingPeerResponse)
	if !ok {
		e := fmt.Errorf("%s. could not convert a response to type GetAddingPeerResponse", logger.Prefix())
		return GetAddingPeerResponse{false, e.Error()}, e
	}
	return *response, nil
}

// create a inner variable in order to mock it in test
var httpClient *http.Client = http.DefaultClient

// ==========
func getReq(ctx context.Context, p *PeerNode, path string, resp any) (any, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/%s", p.TcpAddressWithProtocol(), path), nil)
	if err != nil {
		return nil, err
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return &resp, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("Not found")
	}
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return resp, err
	}
	return resp, nil
}
