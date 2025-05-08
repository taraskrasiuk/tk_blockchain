package main

import (
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/block"
	"taraskrasiuk/blockchain_l/internal/state"
	"taraskrasiuk/blockchain_l/internal/transactions"
)

func main() {
	s := state.NewState()
	defer s.Close()

	block0 := block.NewBlock(block.Hash{}, []transactions.Tx{
		*transactions.NewTx(transactions.Account("andrej"), transactions.Account("andrej"), "", 3),
		*transactions.NewTx(transactions.Account("andrej"), transactions.Account("andrej"), "reward", 700),
	})

	s.AddBlock(block0)

	block0Hash, err := s.PersistV2()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("block hash: %x\n", block0Hash)
	fmt.Printf("parent block hash: %x\n", block0.Header.ParentHash)

	block1 := block.NewBlock(block0Hash, []transactions.Tx{
		*transactions.NewTx("andrej", "babayaga", "", 2000),
		*transactions.NewTx("andrej", "andrej", "reward", 100),
		*transactions.NewTx("babayaga", "andrej", "", 1),
		*transactions.NewTx("babayaga", "caesar", "", 1000),
		*transactions.NewTx("babayaga", "andrej", "", 50),
		*transactions.NewTx("andrej", "andrej", "reward", 600),
	})

	s.AddBlock(block1)

	block1Hash, err := s.PersistV2()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("block hash: %x\n", block1Hash)
	fmt.Printf("parent block hash: %x\n", block1.Header.ParentHash)
}
