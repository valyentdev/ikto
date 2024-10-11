package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ikto/pkg/ikto"
	"github.com/valyentdev/ikto/pkg/server"
)

func NewAgentCommand() *cobra.Command {
	var configPath string
	var socket string
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start the ikto agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(configPath) == 0 {
				return fmt.Errorf("config path is required")
			}

			configFile, err := os.ReadFile(configPath)
			if err != nil {
				return err
			}

			var config Config
			if err := json.Unmarshal(configFile, &config); err != nil {
				return err
			}

			iktoConfig, err := config.Validate()
			if err != nil {
				return err
			}

			ikto, err := ikto.NewIkto(&iktoConfig)
			if err != nil {
				return err
			}

			err = ikto.Start()
			if err != nil {
				return err
			}

			defer ikto.Stop()

			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

			server.StartAdminServer(ctx, *ikto, socket)

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to the configuration file")
	cmd.MarkFlagRequired("config")

	cmd.Flags().StringVarP(&socket, "socket", "s", "/tmp/ikto.sock", "Path to the socket file")

	return cmd
}
