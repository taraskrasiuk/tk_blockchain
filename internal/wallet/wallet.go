package wallet

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

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
