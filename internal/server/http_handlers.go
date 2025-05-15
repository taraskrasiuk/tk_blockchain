package server

import (
	"encoding/json"
	"net/http"
	"taraskrasiuk/blockchain_l/internal/database"
	"taraskrasiuk/blockchain_l/internal/node"
)

type HttpNodeHandler struct {
	node *node.Node
}

func NewHttpNodeHanlder(n *node.Node) *HttpNodeHandler {
	return &HttpNodeHandler{n}
}

// ===== GET /health
func (h *HttpNodeHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ===== GET /balances/list
func (h *HttpNodeHandler) handleGetBalancesList(w http.ResponseWriter, r *http.Request) {
	res := h.node.ViewBalancesList()
	if err := writeJSON(w, http.StatusOK, res); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not write response data")
	}
}

// ====== GET /node/status
func (h *HttpNodeHandler) handlerNodeStatus(w http.ResponseWriter, r *http.Request) {
	resp := h.node.ViewNodeStatus()
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not write response data")
	}
}

// ====== GET /node/sync?fromBlock=xxx
func (h *HttpNodeHandler) handlerSync(w http.ResponseWriter, r *http.Request) {
	reqHash := r.URL.Query().Get("fromBlock")
	if reqHash == "" {
		writeErr(w, http.StatusBadRequest, "fromBlock parameter not found")
		return
	}
	hash := database.Hash{}
	err := hash.UnmarshalText([]byte(reqHash))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "could not validate a provided hash")
		return
	}
	blocks, err := h.node.ViewSyncBlocks(hash)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not get the blocks. internal error")
		return
	}
	writeJSON(w, http.StatusOK, blocks)
}

func (h *HttpNodeHandler) handlerTxAddRequest(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Data  string `json:"data"`
		Value uint   `json:"value"`
	}

	var txReqBody reqBody

	if err := json.NewDecoder(r.Body).Decode(&txReqBody); err != nil {
		writeErr(w, http.StatusBadRequest, "could not decode payload")
		return
	}
	defer r.Body.Close()

	hash, err := h.node.AddTransaction(txReqBody.From, txReqBody.To, txReqBody.Data, txReqBody.Value)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "could not create a new transaction due to: "+err.Error())
		return
	}
	resp := struct {
		Hash database.Hash `json:"hash"`
	}{
		Hash: hash,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create a response message due to: "+err.Error())
		return
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

func writeErr(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
