package main

import (
	"fmt"
	"taraskrasiuk/blockchain_l/internal/state"
)

func main() {
	st := state.NewState()

	fmt.Println(st.Balances)
	fmt.Println(st.TxMempool)
}
