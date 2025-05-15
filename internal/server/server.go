package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"taraskrasiuk/blockchain_l/internal/node"
)

type NodeServer struct {
	node   *node.Node
	logger *log.Logger
	port   int
}

func (s *NodeServer) Run(ctx context.Context) error {
	s.logger.Printf(" node server is running on port %d", s.port)

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

	withLogger := NewLoggerMiddleware(mux, os.Stdout)

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), withLogger)
}
