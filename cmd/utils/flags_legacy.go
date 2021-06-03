package utils

import (
	"fmt"
	"strings"

	"github.com/pecoin/go-fullnode/node"
	"gopkg.in/urfave/cli.v1"
)

var ShowDeprecated = cli.Command{
	Action:      showDeprecated,
	Name:        "show-deprecated-flags",
	Usage:       "Show flags that have been deprecated",
	ArgsUsage:   " ",
	Category:    "MISCELLANEOUS COMMANDS",
	Description: "Show flags that have been deprecated and will soon be removed",
}

var DeprecatedFlags = []cli.Flag{}

var (
	// (Deprecated May 2020, shown in aliased flags section)
	LegacyRPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable the HTTP-RPC server (deprecated and will be removed June 2021, use --http)",
	}
	LegacyRPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface (deprecated and will be removed June 2021, use --http.addr)",
		Value: node.DefaultHTTPHost,
	}
	LegacyRPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port (deprecated and will be removed June 2021, use --http.port)",
		Value: node.DefaultHTTPPort,
	}
	LegacyRPCCORSDomainFlag = cli.StringFlag{
		Name:  "rpccorsdomain",
		Usage: "Comma separated list of domains from which to accept cross origin requests (browser enforced) (deprecated and will be removed June 2021, use --http.corsdomain)",
		Value: "",
	}
	LegacyRPCVirtualHostsFlag = cli.StringFlag{
		Name:  "rpcvhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard. (deprecated and will be removed June 2021, use --http.vhosts)",
		Value: strings.Join(node.DefaultConfig.HTTPVirtualHosts, ","),
	}
	LegacyRPCApiFlag = cli.StringFlag{
		Name:  "rpcapi",
		Usage: "API's offered over the HTTP-RPC interface (deprecated and will be removed June 2021, use --http.api)",
		Value: "",
	}
)

// showDeprecated displays deprecated flags that will be soon removed from the codebase.
func showDeprecated(*cli.Context) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("The following flags are deprecated and will be removed in the future!")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println()
	// TODO remove when there are newly deprecated flags
	fmt.Println("no deprecated flags to show at this time")
	fmt.Println()
}
