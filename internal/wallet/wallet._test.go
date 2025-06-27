package wallet

import (
	// "crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_MessageSign(t *testing.T) {
	// create random priv key
	pk, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(pk)

	// pb
	pb := pk.PublicKey
	pbBytes := elliptic.Marshal(crypto.S256(), pb.X, pb.Y)
	pbBytesHash := crypto.Keccak256(pbBytes[1:])

	account := common.BytesToAddress(pbBytesHash[12:])

	msg := []byte("hello world")

	sig, err := Sign(msg, pk)
	if err != nil {
		t.Fatal(err)
	}

	recoveredPb, err := Verify(msg, sig)
	if err != nil {
		t.Fatal(err)
	}
	recoveredPbBytes := elliptic.Marshal(crypto.S256(), recoveredPb.X, recoveredPb.Y)
	recoveredPbBytesHash := crypto.Keccak256(recoveredPbBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPbBytesHash[12:])
	if account.Hex() != recoveredAccount.Hex() {
		t.Fatalf("msg was signed by account %s but signature recovery produced an account %s",
			account.Hex(),
			recoveredAccount.Hex())
	}
}
