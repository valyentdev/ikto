package commands

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var root = &cobra.Command{
		Use:   "ikto",
		Short: "A NATS based wireguard mesh network builder",
		Long:  "A NATS based wireguard mesh network builder",
	}

	root.AddCommand(NewAgentCommand())
	root.AddCommand(NewInitCommand())
	root.AddCommand(NewInfoCommand())
	return root
}
