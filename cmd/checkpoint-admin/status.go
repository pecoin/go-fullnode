package main

import (
	"fmt"

	"github.com/pecoin/go-fullnode/cmd/utils"
	"github.com/pecoin/go-fullnode/common"
	"gopkg.in/urfave/cli.v1"
)

var commandStatus = cli.Command{
	Name:  "status",
	Usage: "Fetches the signers and checkpoint status of the oracle contract",
	Flags: []cli.Flag{
		nodeURLFlag,
	},
	Action: utils.MigrateFlags(status),
}

// status fetches the admin list of specified registrar contract.
func status(ctx *cli.Context) error {
	// Create a wrapper around the checkpoint oracle contract
	addr, oracle := newContract(newRPCClient(ctx.GlobalString(nodeURLFlag.Name)))
	// fmt.Printf("Oracle => %s\n", addr.Hex())
	fmt.Printf("Oracle => %s\n", addr.Base58())
	fmt.Println()

	// Retrieve the list of authorized signers (admins)
	admins, err := oracle.Contract().GetAllAdmin(nil)
	if err != nil {
		return err
	}
	for i, admin := range admins {
		// fmt.Printf("Admin %d => %s\n", i+1, admin.Hex())
		fmt.Printf("Admin %d => %s\n", i+1, admin.Base58())
	}
	fmt.Println()

	// Retrieve the latest checkpoint
	index, checkpoint, height, err := oracle.Contract().GetLatestCheckpoint(nil)
	if err != nil {
		return err
	}
	fmt.Printf("Checkpoint (published at #%d) %d => %s\n", height, index, common.Hash(checkpoint).Hex())

	return nil
}
