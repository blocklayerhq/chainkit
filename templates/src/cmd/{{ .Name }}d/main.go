package main

import (
	"encoding/json"
	"io"
	"os"

	app "{{ .GoPkg }}"
	gaiaInit "github.com/cosmos/cosmos-sdk/cmd/gaia/init"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

// DefaultNodeHome fixme
var DefaultNodeHome = os.ExpandEnv("$HOME/.{{ .Name }}d")

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "{{ .Name }}d",
		Short:             "{{ .Name }} App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	appInit := server.DefaultAppInit
	rootCmd.AddCommand(gaiaInit.InitCmd(ctx, cdc, appInit))
	rootCmd.AddCommand(gaiaInit.TestnetFilesCmd(ctx, cdc, appInit))

	server.AddCommands(ctx, cdc, rootCmd, appInit,
		newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "MA", DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewMyApp(logger, db)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	return nil, nil, nil
}
