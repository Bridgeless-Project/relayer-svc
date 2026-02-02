package service

import (
	"github.com/Bridgeless-Project/relayer-svc/cmd/service/migrate"
	"github.com/Bridgeless-Project/relayer-svc/cmd/service/run"
	"github.com/Bridgeless-Project/relayer-svc/cmd/utils"
	"github.com/spf13/cobra"
)

func init() {
	registerServiceCommands(Cmd)
	utils.RegisterConfigFlag(Cmd)
	utils.RegisterCatchUpFlag(Cmd)
	utils.RegisterStartHeightFlag(Cmd)
}

func registerServiceCommands(cmd *cobra.Command) {
	cmd.AddCommand(migrate.Cmd)
	cmd.AddCommand(run.Cmd)
}

var Cmd = &cobra.Command{
	Use:   "service",
	Short: "Command for running service operations",
}
