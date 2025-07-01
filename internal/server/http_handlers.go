package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"taraskrasiuk/blockchain_l/internal/database"
	"taraskrasiuk/blockchain_l/internal/node"
	"taraskrasiuk/blockchain_l/internal/wallet"
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

// ===== POST /tx/add
func (h *HttpNodeHandler) handlerTxAddRequest(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		From    string `json:"from"`
		FromPWD string `json:"from_pwd"`
		To      string `json:"to"`
		Data    string `json:"data"`
		Value   uint   `json:"value"`
	}

	var txReqBody reqBody

	if err := json.NewDecoder(r.Body).Decode(&txReqBody); err != nil {
		writeErr(w, http.StatusBadRequest, "could not decode payload")
		return
	}
	defer r.Body.Close()

	fromAcc := database.NewAccount(txReqBody.From)
	tx := database.NewTx(fromAcc, database.NewAccount(txReqBody.To), txReqBody.Data, txReqBody.Value, h.node.NextAccountNonce(fromAcc))
	txHash, err := tx.Hash()
	if err != nil {
		writeErr(w, http.StatusBadRequest, "could not create a new pending transcation"+err.Error())
		return
	}
	// create a signed transaction
	fmt.Println(txReqBody.FromPWD, h.node.Dirname())
	signedTx, err := wallet.SignTxWithKeystoreAccount(*tx, database.NewAccount(txReqBody.From), txReqBody.FromPWD, wallet.GetKeystoreDirPath(h.node.Dirname()))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "could not sign a new transaction due to: "+err.Error())
		return
	}
	if err := h.node.AddPendingTX(signedTx); err != nil {
		writeErr(w, http.StatusBadRequest, "could not create a new transaction due to: "+err.Error())
		return
	}
	resp := struct {
		Hash *database.Hash `json:"hash"`
	}{
		Hash: &txHash,
	}
	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create a response message due to: "+err.Error())
		return
	}
}

// ==== GET /node/addpeer?ip=xxxx&port=xxxx
func (h *HttpNodeHandler) handlerAddPeer(w http.ResponseWriter, r *http.Request) {
	var (
		ip   = r.URL.Query().Get("ip")
		port = r.URL.Query().Get("port")
	)
	if ip == "" || port == "" {
		writeErr(w, http.StatusBadRequest, "ip and port should be defined in query")
		return
	}

	peerPort, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	p := node.NewPeerNode(ip, uint(peerPort), false, true)
	h.node.AddPeer(p)
	fmt.Printf("Peer node %s, successfully added.", p.TcpAddress())
	type successRes struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	writeJSON(w, http.StatusOK, &successRes{true, ""})
}

// WALLET
// ==== GET /wallet/accounts
func (h *HttpNodeHandler) handlerWalletAccounts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.node.WalletAccounts())
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
