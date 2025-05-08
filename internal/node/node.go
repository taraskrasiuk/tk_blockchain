package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"taraskrasiuk/blockchain_l/internal/block"
	"taraskrasiuk/blockchain_l/internal/state"
	"taraskrasiuk/blockchain_l/internal/transactions"
)

type BalancesListResponse struct {
	Hash     *block.Hash                   `json:"hash"`
	Balances map[transactions.Account]uint `json:"balances"`
}

type HttpNodeHandler struct {
	s *state.State
}

func NewHttpNodeHanlder(s *state.State) *HttpNodeHandler {
	return &HttpNodeHandler{s}
}

func (h *HttpNodeHandler) handlerBalancesList(w http.ResponseWriter, r *http.Request) {
	balancesListResponse := BalancesListResponse{
		Hash:     &h.s.LastBlockHash,
		Balances: h.s.Balances,
	}

	if err := writeJSON(w, http.StatusOK, balancesListResponse); err != nil {
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
	if err := h.s.Add(*newTx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	hash, err := h.s.Persist()
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

func Run(datadir string) error {
	s, err := state.NewState(datadir)
	if err != nil {
		return err
	}
	defer s.Close()

	// run the http server
	mux := http.NewServeMux()
	nodeHandler := NewHttpNodeHanlder(s)

	mux.HandleFunc("GET /balances/list", nodeHandler.handlerBalancesList)
	mux.HandleFunc("POST /tx/add", nodeHandler.handlerTxAddRequest)

	fmt.Println("A node is running on port 8080")
	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		return err
	}
	return nil
}
