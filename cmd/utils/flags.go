package utils

import (
	"strconv"

	"github.com/Bridgeless-Project/relayer-svc/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/distributed_lab/kit/kv"
)

const (
	configFlag        = "config"
	catchUpFlag       = "catch-up"
	startHeightFlag   = "start-height"
	blockDistanceFlag = "block-distance"
	observerFlag      = "observer"
)

func RegisterConfigFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(configFlag, "c", "config.yaml", "Path to the config file")
}

func RegisterCatchUpFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(catchUpFlag, "u", false, "Catch up unprocessed deposits from database")
}

func RegisterStartHeightFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Uint64P(startHeightFlag, "s", 0, "Start height to fetch blocks")
}

func RegisterBlockDistanceFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Uint64P(blockDistanceFlag, "d", 0, "Block distance between current block and core block")
}

func RegisterObserverFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(observerFlag, "o", false, "Observer address")
}

func ConfigFromFlags(cmd *cobra.Command) (config.Config, error) {
	configPath, err := cmd.Flags().GetString(configFlag)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config flag")
	}

	viper := kv.NewViperFile(configPath)
	if _, err = viper.GetStringMap("ping"); err != nil {
		return nil, errors.Wrap(err, "failed to ping viper")
	}

	return config.New(viper), nil
}

func CatchUpFromFlags(cmd *cobra.Command) (bool, error) {
	catchUp, err := cmd.Flags().GetBool(catchUpFlag)
	if err != nil {
		return false, errors.Wrap(err, "failed to get catch-up flag")
	}

	return catchUp, nil
}

func StartHeightFromFlags(cmd *cobra.Command) (uint64, error) {
	return strconv.ParseUint(cmd.Flags().Lookup(startHeightFlag).Value.String(), 10, 64)
}

func BlockDistanceFromFlags(cmd *cobra.Command) (uint64, error) {
	return strconv.ParseUint(cmd.Flags().Lookup(blockDistanceFlag).Value.String(), 10, 64)
}

func ObserverFromFlags(cmd *cobra.Command) (bool, error) {
	observer, err := cmd.Flags().GetBool(observerFlag)
	if err != nil {
		return false, errors.Wrap(err, "failed to get observer flag")
	}

	return observer, nil
}
