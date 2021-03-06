package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/pecoin/go-fullnode/accounts/keystore"
	"github.com/pecoin/go-fullnode/cmd/utils"
	"github.com/pecoin/go-fullnode/crypto"
	"gopkg.in/urfave/cli.v1"
)

type outputInspect struct {
	Address    string
	PublicKey  string
	PrivateKey string
}

var commandInspect = cli.Command{
	Name:      "inspect",
	Usage:     "inspect a keyfile",
	ArgsUsage: "<keyfile>",
	Description: `
Print various information about the keyfile.

Private key information can be printed by using the --private flag;
make sure to use this feature with great caution!`,
	Flags: []cli.Flag{
		passphraseFlag,
		jsonFlag,
		cli.BoolFlag{
			Name:  "private",
			Usage: "include the private key in the output",
		},
	},
	Action: func(ctx *cli.Context) error {
		keyfilepath := ctx.Args().First()

		// Read key from file.
		keyjson, err := ioutil.ReadFile(keyfilepath)
		if err != nil {
			utils.Fatalf("Failed to read the keyfile at '%s': %v", keyfilepath, err)
		}

		// Decrypt key with passphrase.
		passphrase := getPassphrase(ctx, false)
		key, err := keystore.DecryptKey(keyjson, passphrase)
		if err != nil {
			utils.Fatalf("Error decrypting key: %v", err)
		}

		// Output all relevant information we can retrieve.
		showPrivate := ctx.Bool("private")
		out := outputInspect{
			// Address: key.Address.Hex(),
			Address:    key.Address.Base58(),
			PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(&key.PrivateKey.PublicKey)),
			PrivateKey: hex.EncodeToString(crypto.FromECDSA(key.PrivateKey)),
		}
		if showPrivate {
			out.PrivateKey = hex.EncodeToString(crypto.FromECDSA(key.PrivateKey))
		}

		if ctx.Bool(jsonFlag.Name) {
			mustPrintJSON(out)
		} else {
			fmt.Println("Address:       ", out.Address)
			fmt.Println("Public key:    ", out.PublicKey)
			fmt.Println("Private key:   ", out.PrivateKey)
			if showPrivate {
				fmt.Println("Private key:   ", out.PrivateKey)
			}
		}
		return nil
	},
}
