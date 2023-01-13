package cmd

import (
	"github.com/angelini/netz/pkg/builder"
	"github.com/angelini/netz/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCmdBuild() *cobra.Command {
	var (
		configPath string
	)

	cmd := &cobra.Command{
		Use:   "build [all | <name>]",
		Short: "Generate all build inputs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			log := ctx.Value(logKey).(*zap.Logger)

			root, err := config.Parse(configPath)
			if err != nil {
				return err
			}

			distDir := "dist"

			if args[0] == "all" {
				err = builder.BuildAllLocal(log, root, distDir)
			} else {
				err = builder.BuildLocal(log, root, args[0], distDir)
			}
			if err != nil {
				return err
			}

			err = builder.BuildFront(log, root, distDir)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.hcl", "Config path")

	cmd.MarkPersistentFlagRequired("name")
	cmd.MarkPersistentFlagFilename("config")

	return cmd
}
