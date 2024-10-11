package main

import (
	"fmt"
	"os"

	"github.com/valyentdev/ikto/pkg/commands"
)

func main() {
	root := commands.NewRootCommand()

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
