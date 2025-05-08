package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	genesisFile = "genesis.db"
	blocksFile  = "blocks.db"
)

func initDbDirStructureIfNotExist(dirname string) error {
	// check for genesis
	if fileExists(getGenesisFile(dirname)) {
		return nil
	}
	path := getDbDir(dirname)
	// create subdirectory
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		return err
	}
	// create geneis file
	if err := writeGenesisFile(path); err != nil {
		return err
	}
	// create blocks db
	if err := writeBlocksDbFile(path); err != nil {
		return err
	}
	return nil
}

func getDbDir(dirname string) string {
	return filepath.Join(dirname, "database")
}

func getGenesisFile(dirname string) string {
	return filepath.Join(dirname, genesisFile)
}

func writeGenesisFile(dirname string) error {
	t := time.Now().Format(time.RFC3339)
	genesisFile := fmt.Sprintf(`{
	"genesis_time": "%s",
	"chain_id": "bb ledger",
	"balances": {
    	"andrej": 1000000
    }
}`, t)
	if err := os.WriteFile(getGenesisFile(dirname), []byte(genesisFile), 0644); err != nil {
		return err
	}
	return nil
}

func getBlocksDbFile(dirname string) string {
	return filepath.Join(dirname, "blocks.db")
}

func writeBlocksDbFile(dirname string) error {
	if err := os.WriteFile(filepath.Join(dirname, blocksFileDb), []byte{}, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
