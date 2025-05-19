package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"taraskrasiuk/blockchain_l/internal/node"
)

type NodeServer struct {
	node *node.Node
	port uint
}

func NewNodeServer(n *node.Node, p uint) *NodeServer {
	return &NodeServer{n, p}
}

func (s *NodeServer) Run(ctx context.Context) error {
	logger.Printf(" node server is running on port %d", s.port)
	if err := s.node.Run(ctx); err != nil {
		return err
	}
	defer s.node.Close()

	mux := http.NewServeMux()
	nodeHandler := NewHttpNodeHanlder(s.node)

	mux.HandleFunc("GET /health", nodeHandler.handleHealthCheck)
	// balances
	mux.HandleFunc("GET /balances/list", nodeHandler.handleGetBalancesList)
	// add new transaction
	mux.HandleFunc("POST /tx/add", nodeHandler.handlerTxAddRequest)
	// node
	mux.HandleFunc("GET /node/status", nodeHandler.handlerNodeStatus)
	mux.HandleFunc("GET /node/sync", nodeHandler.handlerSync)
	mux.HandleFunc("GET /node/addpeer", nodeHandler.handlerAddPeer)

	withLogger := NewLoggerMiddleware(mux, os.Stdout)

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), withLogger)
}
