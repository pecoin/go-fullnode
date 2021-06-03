package keystore

import (
	"os"

	"github.com/pecoin/go-fullnode/accounts/keystore"
)

func Fuzz(input []byte) int {
	ks := keystore.NewKeyStore("/tmp/ks", keystore.LightScryptN, keystore.LightScryptP)

	a, err := ks.NewAccount(string(input))
	if err != nil {
		panic(err)
	}
	if err := ks.Unlock(a, string(input)); err != nil {
		panic(err)
	}
	os.Remove(a.URL.Path)
	return 1
}
