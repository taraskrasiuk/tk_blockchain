package utils

import (
	"encoding/json"
	"io"
	"os"
)

func loadFile[T struct{}](filename string) (*T, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
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
	var genesisData T
	err = json.Unmarshal(res, &genesisData)
	if err != nil {
		return nil, err
	}

	return &genesisData, nil

	// result := make(map[transactions.Account]uint)
	// // set a balances to state
	// for k, v := range genesisData.Balances {
	// 	result[transactions.Account(k)] = v
	// }

	// return result, nil
}
