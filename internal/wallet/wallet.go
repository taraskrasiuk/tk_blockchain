package wallet

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"taraskrasiuk/blockchain_l/internal/database"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const keydir = "keystore"

func Sign(msg []byte, privKey *ecdsa.PrivateKey) ([]byte, error) {
	// create a hashed messaged to 32 byte
	msgHash := sha256.Sum256(msg)

	sig, err := crypto.Sign(msgHash[:], privKey)
	if err != nil {
		return nil, err
	}
	if len(sig) != crypto.SignatureLength {
		return nil, fmt.Errorf(
			"wrong size for signature: got %d, want %d",
			len(sig),
			crypto.SignatureLength,
		)
	}
	return sig, err
}

func Verify(msg, sig []byte) (*ecdsa.PublicKey, error) {
	hashedMsg := sha256.Sum256(msg)
	pb, err := crypto.SigToPub(hashedMsg[:], sig)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

// Sign a transaction
func SignTx(tx database.Tx, pk *ecdsa.PrivateKey) (database.SignedTx, error) {
	rawTx, err := tx.Encode()
	if err != nil {
		return database.SignedTx{}, err
	}

	sig, err := Sign(rawTx, pk)
	if err != nil {
		return database.SignedTx{}, err
	}
	return *database.NewSignedTx(tx, sig), nil
}

// Helper, sign a transaction based on provided password and keystore directory, in order to find an account
func SignTxWithKeystoreAccount(tx database.Tx, acc database.Account, pass, keydir string) (database.SignedTx, error) {
	ks := keystore.NewKeyStore(keydir, keystore.StandardScryptN, keystore.StandardScryptP)
	foundedAcc, err := ks.Find(accounts.Account{Address: common.Address(acc)})
	if err != nil {
		return database.SignedTx{}, err
	}
	accJson, err := os.ReadFile(foundedAcc.URL.Path)
	if err != nil {
		return database.SignedTx{}, err
	}
	fmt.Println("ACC JSON: \n" + string(accJson))
	pk, err := keystore.DecryptKey(accJson, pass)
	if err != nil {
		return database.SignedTx{}, err
	}
	signedTx, err := SignTx(tx, pk.PrivateKey)
	if err != nil {
		return database.SignedTx{}, err
	}
	return signedTx, nil
}

func GetKeystoreDirPath(dataDir string) string {
	return filepath.Join(dataDir, keydir)
}
