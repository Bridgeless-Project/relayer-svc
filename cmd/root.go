package cmd

import (
	"os"

	"github.com/Bridgeless-Project/relayer-svc/cmd/service"
	"github.com/spf13/cobra"
)

func Execute() {
	root := &cobra.Command{
		Use:   "relayer-svc",
		Short: "Relayer service",
	}

	root.AddCommand(service.Cmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
