package database

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

var (
	defaultChainID = "123"
)

type GenesisResource struct {
	GenesisTime string          `json:"genesis_time"`
	ChainID     string          `json:"chain_id"`
	Balances    map[string]uint `json:"balances"`
}

func NewGenesisResource() *GenesisResource {
	return &GenesisResource{
		GenesisTime: time.Now().Format(time.RFC3339),
		ChainID:     defaultChainID,
		Balances:    make(map[string]uint),
	}
}

func (g *GenesisResource) AddAccount(hexAddr string, balance uint) {
	g.Balances[hexAddr] = balance
}

func (g *GenesisResource) SaveToFile(filepath string) error {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(g); err != nil {
		return err
	}
	return nil
}

func (g *GenesisResource) LoadFromFile(filepath string) error {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 128)
	res := []byte{}
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		res = append(res, buf[:n]...)
	}

	var genesisData GenesisResource
	err = json.Unmarshal(res, &genesisData)
	if err != nil {
		return err
	}
	// set a balances to state
	for k, v := range genesisData.Balances {
		g.Balances[k] = v
	}

	// set other fields
	g.ChainID = genesisData.ChainID
	g.GenesisTime = genesisData.GenesisTime

	return nil
}
