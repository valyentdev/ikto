package commands

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var root = &cobra.Command{
		Use:   "ikto",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`,
	}

	root.AddCommand(NewAgentCommand())
	root.AddCommand(NewInitCommand())
	root.AddCommand(NewInfoCommand())
	return root
}
